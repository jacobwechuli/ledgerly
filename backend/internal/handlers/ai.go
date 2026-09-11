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
		store:      s,
		validate:   validator.New(),
		openAIKey:  openAIKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (h *AIHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/chat", h.Chat)
	mux.HandleFunc("/insights", h.Insights)
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

	// Build system prompt with financial context - concierge tone
	systemPrompt := fmt.Sprintf(`You are Ledgerly AI, a sophisticated private wealth concierge assistant serving high-net-worth clients. Your tone is professional, discreet, and insightful — like a senior relationship manager at a private bank.

You provide clear, data-driven answers about the client's finances. You never use gamified language, emojis, or motivational platitudes. You speak with authority and precision.

Here is the client's current financial position:
%s

When answering questions:
- Be specific with numbers and percentages
- Reference actual account balances and transaction data
- Provide context for changes (e.g., "Your USD exposure decreased 8% this month due to...")
- Offer strategic observations, not generic advice
- Maintain discretion — never suggest the client is "doing well" or "needs improvement" in a judgmental way
- Focus on actionable insights and risk assessment`, financialContext)

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

// Insights generates concierge-style observations about the user's finances
func (h *AIHandler) Insights(w http.ResponseWriter, r *http.Request) {
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

	// Get net worth summary
	netWorth, err := h.store.GetNetWorthSummary(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch net worth"})
		return
	}

	// Get currency exposure
	currencyExposure, err := h.store.GetCurrencyExposure(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to analyze currency exposure"})
		return
	}

	// Estimate capital gains
	capitalGains, err := h.store.EstimateCapitalGains(r.Context(), userID, time.Now().Year())
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to estimate capital gains"})
		return
	}

	prompt := fmt.Sprintf(`You are a senior private wealth advisor generating a concise financial briefing for a high-net-worth client. Analyze the following data and generate 3-5 key insights that a relationship manager would highlight in a quarterly review.

Each insight should be:
- Specific and data-driven (use actual numbers)
- Actionable or risk-aware
- Professional in tone (no emojis, no gamification, no motivational language)
- Categorized as: portfolio, cash_flow, currency, tax, or goals
- Rated by priority: high, medium, or low

Format your response as a JSON array of objects with these fields:
- "category": one of "portfolio", "cash_flow", "currency", "tax", "goals"
- "title": brief headline (max 8 words)
- "description": detailed observation (2-3 sentences)
- "impact": "positive", "negative", or "neutral"
- "priority": "high", "medium", or "low"

Financial Data:
%s

Net Worth Summary:
- Total: $%.2f USD
- Cash: $%.2f USD
- Investments: $%.2f USD
- Property: $%.2f USD
- 30-day change: %.2f%%

Currency Exposure (last 30 days):
%v

Unrealized Capital Gains: $%.2f USD

Generate insights that a sophisticated client would find valuable. Focus on:
- Portfolio concentration risks
- Currency exposure changes
- Tax implications of unrealized gains
- Goal progress relative to deadlines
- Cash flow patterns and anomalies`, financialContext, netWorth.TotalUSD, netWorth.CashUSD, netWorth.InvestmentsUSD, netWorth.PropertyUSD, netWorth.ChangePct30d, currencyExposure, capitalGains)

	messages := []map[string]string{
		{"role": "system", "content": "You are a private wealth advisor. Always respond with valid JSON when asked."},
		{"role": "user", "content": prompt},
	}

	response, err := h.callOpenAI(r.Context(), messages)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "AI service error", Message: err.Error()})
		return
	}

	// Try to parse the AI response as an array of insights
	var insights []models.AIInsight
	if err := json.Unmarshal([]byte(response), &insights); err != nil {
		// If parsing fails, return the raw response
		respondJSON(w, http.StatusOK, models.APIResponse{Data: map[string]interface{}{
			"raw_response": response,
			"generated_at": time.Now(),
		}})
		return
	}

	// Add IDs and timestamps
	for i := range insights {
		insights[i].ID = fmt.Sprintf("insight_%d", i+1)
		insights[i].GeneratedAt = time.Now()
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: insights})
}

func (h *AIHandler) buildFinancialContext(ctx context.Context, userID string) (string, error) {
	// Get accounts
	accounts, err := h.store.GetAccounts(ctx, userID)
	if err != nil {
		return "", err
	}

	// Get recent transactions
	transactions, _, err := h.store.GetTransactions(ctx, userID, 1, 30)
	if err != nil {
		return "", err
	}

	// Get investments
	investments, err := h.store.GetInvestments(ctx, userID)
	if err != nil {
		return "", err
	}

	// Get properties
	properties, err := h.store.GetProperties(ctx, userID)
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

	var context bytes.Buffer

	context.WriteString("ACCOUNTS:\n")
	for _, a := range accounts {
		context.WriteString(fmt.Sprintf("- %s (%s, %s): %.2f %s\n",
			a.Name, a.Type, a.SubType, a.Balance, a.Currency))
	}

	context.WriteString("\nRECENT TRANSACTIONS (last 30):\n")
	for _, t := range transactions {
		context.WriteString(fmt.Sprintf("- %s: %s %.2f %s (%s) on %s\n",
			t.Type, t.Description, t.Amount, t.Currency, t.Category, t.Date.Format("2006-01-02")))
	}

	context.WriteString("\nINVESTMENTS:\n")
	for _, i := range investments {
		context.WriteString(fmt.Sprintf("- %s (%s): %.4f units @ %.2f %s, unrealized gain: %.2f %s\n",
			i.Symbol, i.Type, i.Quantity, i.CurrentPrice, i.Currency, i.UnrealizedGain, i.Currency))
	}

	context.WriteString("\nPROPERTIES:\n")
	for _, p := range properties {
		context.WriteString(fmt.Sprintf("- %s (%s): %.2f %s (purchased: %.2f %s)\n",
			p.Name, p.Type, p.CurrentValue, p.Currency, p.PurchasePrice, p.Currency))
	}

	context.WriteString("\nBUDGET GOALS:\n")
	for _, g := range goals {
		progress := 0.0
		if g.TargetAmount > 0 {
			progress = (g.CurrentAmount / g.TargetAmount) * 100
		}
		context.WriteString(fmt.Sprintf("- %s: %.2f/%.2f %s (%.1f%%) - %s priority\n",
			g.Name, g.CurrentAmount, g.TargetAmount, g.Currency, progress, g.Priority))
	}

	context.WriteString("\nUPCOMING BILLS:\n")
	for _, b := range bills {
		context.WriteString(fmt.Sprintf("- %s: %.2f %s due %s (%s)\n",
			b.Name, b.Amount, b.Currency, b.DueDate.Format("2006-01-02"), b.Recurrence))
	}

	return context.String(), nil
}

func (h *AIHandler) callOpenAI(ctx context.Context, messages []map[string]string) (string, error) {
	if h.openAIKey == "" {
		return "AI features are not configured. Please set OPENAI_API_KEY.", nil
	}

	requestBody := map[string]interface{}{
		"model":       "gpt-4o",
		"messages":    messages,
		"max_tokens":  1500,
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
