package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/ledgerly/backend/internal/auth"
	"github.com/ledgerly/backend/internal/models"
	"github.com/ledgerly/backend/internal/store"
)

type AIHandler struct {
	store      *store.Store
	validate   *validator.Validate
	openAIKey  string
	httpClient *http.Client
}

func NewAIHandler(s *store.Store, openAIKey string) *AIHandler {
	return &AIHandler{
		store:    s,
		validate: validator.New(),
		openAIKey: openAIKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (h *AIHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/chat", h.Chat)
	mux.HandleFunc("/suggest-budget", h.SuggestBudget)
	return mux
}

// Chat handles AI chat requests about the user's finances
func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "method not allowed"})
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	var req models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "validation failed", Message: err.Error()})
		return
	}

	// Fetch user's financial context
	financialContext, err := h.buildFinancialContext(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to build financial context"})
		return
	}

	// Build system prompt with financial context
	systemPrompt := fmt.Sprintf(`You are Ledgerly AI, a helpful personal finance assistant. You help users understand their finances, track spending, and make better financial decisions.

Here is the user's current financial context:
%s

Answer the user's questions based on their financial data. Be specific with numbers when relevant. If you don't have enough data to answer accurately, say so. Always be encouraging and helpful about financial management.`, financialContext)

	// Build messages for OpenAI
	messages := []map[string]string{
		{"role": "system", "content": systemPrompt},
	}

	for _, msg := range req.Messages {
		messages = append(messages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// Call OpenAI
	response, err := h.callOpenAI(r.Context(), messages)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "AI service error", Message: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: map[string]string{
		"response": response,
	}})
}

// SuggestBudget generates budget recommendations from transaction history
func (h *AIHandler) SuggestBudget(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, models.APIError{Error: "method not allowed"})
		return
	}

	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	// Get user's financial data
	financialContext, err := h.buildFinancialContext(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to analyze finances"})
		return
	}

	totalIncome, totalExpense, err := h.store.GetTransactionSummary(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to get summary"})
		return
	}

	prompt := fmt.Sprintf(`Based on the following financial data, suggest a monthly budget allocation. 
Return your response as a JSON object with this exact structure:
{
  "categories": [
    {"category": "category_name", "recommended": amount, "percentage": pct, "rationale": "why this allocation"}
  ],
  "notes": "general budget advice for this user"
}

Financial data:
%s

Total monthly income: %.2f
Total monthly expenses: %.2f

Suggest allocations for common categories like: housing, food, transport, entertainment, savings, utilities, healthcare, education.
Make sure percentages add up to 100 and recommended amounts are based on the income level.`, financialContext, totalIncome, totalExpense)

	messages := []map[string]string{
		{"role": "system", "content": "You are a financial advisor AI. Always respond with valid JSON when asked."},
		{"role": "user", "content": prompt},
	}

	response, err := h.callOpenAI(r.Context(), messages)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "AI service error", Message: err.Error()})
		return
	}

	// Try to parse the AI response as a BudgetSuggestion
	var suggestion models.BudgetSuggestion
	if err := json.Unmarshal([]byte(response), &suggestion); err != nil {
		// If parsing fails, return the raw response
		respondJSON(w, http.StatusOK, models.APIResponse{Data: map[string]interface{}{
			"raw_response": response,
			"total_income": totalIncome,
			"total_expense": totalExpense,
		}})
		return
	}

	suggestion.TotalIncome = totalIncome
	respondJSON(w, http.StatusOK, models.APIResponse{Data: suggestion})
}

func (h *AIHandler) buildFinancialContext(ctx context.Context, userID string) (string, error) {
	// Get recent transactions
	transactions, _, err := h.store.GetTransactions(ctx, userID, 1, 50)
	if err != nil {
		return "", err
	}

	// Get budget goals
	goals, err := h.store.GetBudgetGoals(ctx, userID)
	if err != nil {
		return "", err
	}

	// Get upcoming bills
	bills, err := h.store.GetBills(ctx, userID, boolPtr(false))
	if err != nil {
		return "", err
	}

	// Get expenditure report for last 30 days
	thirtyDaysAgo := time.Now().AddDate(0, -1, 0)
	report, err := h.store.GetExpenditureReport(ctx, userID, thirtyDaysAgo, time.Now())
	if err != nil {
		return "", err
	}

	var context bytes.Buffer

	context.WriteString("Recent Transactions (last 50):\n")
	for _, t := range transactions {
		context.WriteString(fmt.Sprintf("- %s: %s %.2f (%s, %s) on %s\n",
			t.Type, t.Description, t.Amount, t.Category, t.Source, t.Date.Format("2006-01-02")))
	}

	context.WriteString("\nBudget Goals:\n")
	for _, g := range goals {
		progress := 0.0
		if g.TargetAmount > 0 {
			progress = (g.CurrentAmount / g.TargetAmount) * 100
		}
		context.WriteString(fmt.Sprintf("- %s: %.2f/%.2f (%.1f%% complete) - %s\n",
			g.Name, g.CurrentAmount, g.TargetAmount, progress, g.Status))
	}

	context.WriteString("\nUpcoming Bills:\n")
	for _, b := range bills {
		context.WriteString(fmt.Sprintf("- %s: %.2f due %s (%s)\n",
			b.Name, b.Amount, b.DueDate.Format("2006-01-02"), b.Recurrence))
	}

	context.WriteString("\nSpending by Category (last 30 days):\n")
	for _, cs := range report.Categories {
		context.WriteString(fmt.Sprintf("- %s: %.2f (%d transactions)\n",
			cs.Category, cs.Amount, cs.Count))
	}
	context.WriteString(fmt.Sprintf("Total spent: %.2f\n", report.TotalSpent))

	return context.String(), nil
}

func (h *AIHandler) callOpenAI(ctx context.Context, messages []map[string]string) (string, error) {
	if h.openAIKey == "" {
		return "AI features are not configured. Please set OPENAI_API_KEY.", nil
	}

	requestBody := map[string]interface{}{
		"model":    "gpt-4o-mini",
		"messages": messages,
		"max_tokens": 1000,
		"temperature": 0.7,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.openAIKey)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

func boolPtr(b bool) *bool {
	return &b
}
