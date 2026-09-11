package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/ledgerly/backend/internal/auth"
	"github.com/ledgerly/backend/internal/models"
	"github.com/ledgerly/backend/internal/store"
)

// --- Account Handler ---

type AccountHandler struct {
	store    *store.Store
	validate *validator.Validate
}

func NewAccountHandler(s *store.Store) *AccountHandler {
	return &AccountHandler{store: s, validate: validator.New()}
}

func (h *AccountHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Put("/{id}/balance", h.UpdateBalance)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	var req models.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "validation failed", Message: err.Error()})
		return
	}

	account, err := h.store.CreateAccount(r.Context(), userID, req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to create account"})
		return
	}

	respondJSON(w, http.StatusCreated, models.APIResponse{Data: account})
}

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	accounts, err := h.store.GetAccounts(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch accounts"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: accounts})
}

func (h *AccountHandler) UpdateBalance(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid account ID"})
		return
	}

	var body struct {
		Balance float64 `json:"balance" validate:"gte=0"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid request body"})
		return
	}

	if err := h.store.UpdateAccountBalance(r.Context(), userID, id, body.Balance); err != nil {
		respondJSON(w, http.StatusNotFound, models.APIError{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Message: "balance updated"})
}

func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid account ID"})
		return
	}

	if err := h.store.DeleteAccount(r.Context(), userID, id); err != nil {
		respondJSON(w, http.StatusNotFound, models.APIError{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Message: "account deactivated"})
}

// --- Transaction Handler ---

type TransactionHandler struct {
	store    *store.Store
	validate *validator.Validate
}

func NewTransactionHandler(s *store.Store) *TransactionHandler {
	return &TransactionHandler{store: s, validate: validator.New()}
}

func (h *TransactionHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	var req models.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "validation failed", Message: err.Error()})
		return
	}

	tx, err := h.store.CreateTransaction(r.Context(), userID, req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to create transaction"})
		return
	}

	respondJSON(w, http.StatusCreated, models.APIResponse{Data: tx})
}

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	transactions, totalCount, err := h.store.GetTransactions(r.Context(), userID, page, pageSize)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch transactions"})
		return
	}

	respondJSON(w, http.StatusOK, models.PaginatedResponse{
		Data:       transactions,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: store.Paginate(totalCount, pageSize),
	})
}

func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid transaction ID"})
		return
	}

	if err := h.store.DeleteTransaction(r.Context(), userID, id); err != nil {
		respondJSON(w, http.StatusNotFound, models.APIError{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Message: "transaction deleted"})
}

// --- Investment Handler ---

type InvestmentHandler struct {
	store    *store.Store
	validate *validator.Validate
}

func NewInvestmentHandler(s *store.Store) *InvestmentHandler {
	return &InvestmentHandler{store: s, validate: validator.New()}
}

func (h *InvestmentHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *InvestmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	var req models.CreateInvestmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "validation failed", Message: err.Error()})
		return
	}

	inv, err := h.store.CreateInvestment(r.Context(), userID, req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to create investment"})
		return
	}

	respondJSON(w, http.StatusCreated, models.APIResponse{Data: inv})
}

func (h *InvestmentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	investments, err := h.store.GetInvestments(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch investments"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: investments})
}

func (h *InvestmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid investment ID"})
		return
	}

	if err := h.store.DeleteInvestment(r.Context(), userID, id); err != nil {
		respondJSON(w, http.StatusNotFound, models.APIError{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Message: "investment deleted"})
}

// --- Property Handler ---

type PropertyHandler struct {
	store    *store.Store
	validate *validator.Validate
}

func NewPropertyHandler(s *store.Store) *PropertyHandler {
	return &PropertyHandler{store: s, validate: validator.New()}
}

func (h *PropertyHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *PropertyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	var req models.CreatePropertyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "validation failed", Message: err.Error()})
		return
	}

	prop, err := h.store.CreateProperty(r.Context(), userID, req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to create property"})
		return
	}

	respondJSON(w, http.StatusCreated, models.APIResponse{Data: prop})
}

func (h *PropertyHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	properties, err := h.store.GetProperties(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch properties"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: properties})
}

func (h *PropertyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid property ID"})
		return
	}

	if err := h.store.DeleteProperty(r.Context(), userID, id); err != nil {
		respondJSON(w, http.StatusNotFound, models.APIError{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Message: "property deleted"})
}

// --- Net Worth Handler ---

type NetWorthHandler struct {
	store *store.Store
}

func NewNetWorthHandler(s *store.Store) *NetWorthHandler {
	return &NetWorthHandler{store: s}
}

func (h *NetWorthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/summary", h.Summary)
	r.Get("/history", h.History)
	return r
}

func (h *NetWorthHandler) Summary(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	summary, err := h.store.GetNetWorthSummary(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch net worth"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: summary})
}

func (h *NetWorthHandler) History(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30
	}

	history, err := h.store.GetNetWorthHistory(r.Context(), userID, days)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch history"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: history})
}

// --- Currency Handler ---

type CurrencyHandler struct {
	store *store.Store
}

func NewCurrencyHandler(s *store.Store) *CurrencyHandler {
	return &CurrencyHandler{store: s}
}

func (h *CurrencyHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/convert", h.Convert)
	return r
}

func (h *CurrencyHandler) List(w http.ResponseWriter, r *http.Request) {
	currencies, err := h.store.GetAllCurrencies(r.Context())
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch currencies"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: currencies})
}

func (h *CurrencyHandler) Convert(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	amountStr := r.URL.Query().Get("amount")

	if from == "" || to == "" || amountStr == "" {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "missing required parameters: from, to, amount"})
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid amount"})
		return
	}

	result, err := h.store.ConvertCurrency(r.Context(), from, to, amount)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: result})
}

// --- Bill Handler ---

type BillHandler struct {
	store    *store.Store
	validate *validator.Validate
}

func NewBillHandler(s *store.Store) *BillHandler {
	return &BillHandler{store: s, validate: validator.New()}
}

func (h *BillHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Post("/{id}/pay", h.MarkPaid)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *BillHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	var req models.CreateBillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "validation failed", Message: err.Error()})
		return
	}

	bill, err := h.store.CreateBill(r.Context(), userID, req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to create bill"})
		return
	}

	respondJSON(w, http.StatusCreated, models.APIResponse{Data: bill})
}

func (h *BillHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	var paidFilter *bool
	if paidParam := r.URL.Query().Get("paid"); paidParam != "" {
		paid := paidParam == "true"
		paidFilter = &paid
	}

	bills, err := h.store.GetBills(r.Context(), userID, paidFilter)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch bills"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: bills})
}

func (h *BillHandler) MarkPaid(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid bill ID"})
		return
	}

	bill, err := h.store.MarkBillPaid(r.Context(), userID, id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, models.APIError{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: bill})
}

func (h *BillHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid bill ID"})
		return
	}

	if err := h.store.DeleteBill(r.Context(), userID, id); err != nil {
		respondJSON(w, http.StatusNotFound, models.APIError{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Message: "bill deleted"})
}

// --- Budget Goal Handler ---

type BudgetGoalHandler struct {
	store    *store.Store
	validate *validator.Validate
}

func NewBudgetGoalHandler(s *store.Store) *BudgetGoalHandler {
	return &BudgetGoalHandler{store: s, validate: validator.New()}
}

func (h *BudgetGoalHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Post("/{id}/contribute", h.Contribute)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *BudgetGoalHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	var req models.CreateBudgetGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "validation failed", Message: err.Error()})
		return
	}

	goal, err := h.store.CreateBudgetGoal(r.Context(), userID, req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to create budget goal"})
		return
	}

	respondJSON(w, http.StatusCreated, models.APIResponse{Data: goal})
}

func (h *BudgetGoalHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	goals, err := h.store.GetBudgetGoals(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch budget goals"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: goals})
}

func (h *BudgetGoalHandler) Contribute(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid goal ID"})
		return
	}

	var body struct {
		Amount float64 `json:"amount" validate:"required,gt=0"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid request body"})
		return
	}

	if err := h.validate.Struct(body); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "validation failed", Message: err.Error()})
		return
	}

	goal, err := h.store.UpdateBudgetGoalProgress(r.Context(), userID, id, body.Amount)
	if err != nil {
		respondJSON(w, http.StatusNotFound, models.APIError{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: goal})
}

func (h *BudgetGoalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid goal ID"})
		return
	}

	if err := h.store.DeleteBudgetGoal(r.Context(), userID, id); err != nil {
		respondJSON(w, http.StatusNotFound, models.APIError{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Message: "budget goal deleted"})
}

// --- Budget Template Handler ---

type BudgetTemplateHandler struct {
	store    *store.Store
	validate *validator.Validate
}

func NewBudgetTemplateHandler(s *store.Store) *BudgetTemplateHandler {
	return &BudgetTemplateHandler{store: s, validate: validator.New()}
}

func (h *BudgetTemplateHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *BudgetTemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	var req models.CreateBudgetTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "validation failed", Message: err.Error()})
		return
	}

	template, err := h.store.CreateBudgetTemplate(r.Context(), userID, req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to create budget template"})
		return
	}

	respondJSON(w, http.StatusCreated, models.APIResponse{Data: template})
}

func (h *BudgetTemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	templates, err := h.store.GetBudgetTemplates(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch budget templates"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: templates})
}

func (h *BudgetTemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid template ID"})
		return
	}

	if err := h.store.DeleteBudgetTemplate(r.Context(), userID, id); err != nil {
		respondJSON(w, http.StatusNotFound, models.APIError{Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Message: "template deleted"})
}

// --- Report Handler ---

type ReportHandler struct {
	store *store.Store
}

func NewReportHandler(s *store.Store) *ReportHandler {
	return &ReportHandler{store: s}
}

func (h *ReportHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/expenditure", h.Expenditure)
	r.Get("/summary", h.Summary)
	return r
}

func (h *ReportHandler) Expenditure(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")
	currency := r.URL.Query().Get("currency")

	var startDate, endDate time.Time
	var err error

	if startStr != "" {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid start_date format"})
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0)
	}

	if endStr != "" {
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid end_date format"})
			return
		}
	} else {
		endDate = time.Now()
	}

	report, err := h.store.GetExpenditureReport(r.Context(), userID, startDate, endDate, currency)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to generate report"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: report})
}

func (h *ReportHandler) Summary(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	totalIncome, totalExpense, err := h.store.GetTransactionSummary(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch summary"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: map[string]interface{}{
		"total_income_usd":  totalIncome,
		"total_expense_usd": totalExpense,
		"net_balance_usd":   totalIncome - totalExpense,
	}})
}

// --- User Settings Handler ---

type SettingsHandler struct {
	store    *store.Store
	validate *validator.Validate
}

func NewSettingsHandler(s *store.Store) *SettingsHandler {
	return &SettingsHandler{store: s, validate: validator.New()}
}

func (h *SettingsHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.Get)
	r.Put("/", h.Update)
	return r
}

func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	settings, err := h.store.GetUserSettings(r.Context(), userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch settings"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: settings})
}

func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	var req models.UpdateUserSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid request body"})
		return
	}

	settings, err := h.store.UpdateUserSettings(r.Context(), userID, req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to update settings"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: settings})
}

// --- Tax Handler ---

type TaxHandler struct {
	store *store.Store
}

func NewTaxHandler(s *store.Store) *TaxHandler {
	return &TaxHandler{store: s}
}

func (h *TaxHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/capital-gains", h.CapitalGains)
	r.Get("/records", h.Records)
	return r
}

func (h *TaxHandler) CapitalGains(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	yearStr := r.URL.Query().Get("year")
	year := time.Now().Year()
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	gains, err := h.store.EstimateCapitalGains(r.Context(), userID, year)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to estimate capital gains"})
		return
	}

	// Estimate tax (simplified: 15% long-term capital gains rate)
	estimatedTax := gains * 0.15

	respondJSON(w, http.StatusOK, models.APIResponse{Data: map[string]interface{}{
		"year":               year,
		"unrealized_gains":   gains,
		"estimated_tax_rate": 0.15,
		"estimated_tax":      estimatedTax,
		"currency":           "USD",
		"note":               "Estimate based on 15% long-term capital gains rate. Consult a tax professional for accurate figures.",
	}})
}

func (h *TaxHandler) Records(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	yearStr := r.URL.Query().Get("year")
	year := time.Now().Year()
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	records, err := h.store.GetTaxRecords(r.Context(), userID, year)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch tax records"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: records})
}

// --- Helper ---

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
