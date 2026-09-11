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

type TransactionHandler struct {
	store    *store.Store
	validate *validator.Validate
}

func NewTransactionHandler(s *store.Store) *TransactionHandler {
	return &TransactionHandler{
		store:    s,
		validate: validator.New(),
	}
}

func (h *TransactionHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
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

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	transactions, totalCount, err := h.store.GetTransactions(r.Context(), userID, page, pageSize)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch transactions"})
		return
	}

	totalPages := store.Paginate(totalCount, pageSize)

	respondJSON(w, http.StatusOK, models.PaginatedResponse{
		Data:       transactions,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	})
}

func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	tx, err := h.store.GetTransactionByID(r.Context(), userID, id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, models.APIError{Error: "transaction not found"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Data: tx})
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
		respondJSON(w, http.StatusNotFound, models.APIError{Error: "transaction not found"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Message: "transaction deleted"})
}

// --- Mobile Money Handler ---

type MobileMoneyHandler struct {
	store    *store.Store
	validate *validator.Validate
}

func NewMobileMoneyHandler(s *store.Store) *MobileMoneyHandler {
	return &MobileMoneyHandler{
		store:    s,
		validate: validator.New(),
	}
}

func (h *MobileMoneyHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	return r
}

func (h *MobileMoneyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	var req models.CreateMobileMoneyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid request body"})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.APIError{Error: "validation failed", Message: err.Error()})
		return
	}

	tx, err := h.store.CreateMobileMoneyTransaction(r.Context(), userID, req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to create mobile money transaction"})
		return
	}

	respondJSON(w, http.StatusCreated, models.APIResponse{Data: tx})
}

func (h *MobileMoneyHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	if userID == "" {
		respondJSON(w, http.StatusUnauthorized, models.APIError{Error: "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	txs, totalCount, err := h.store.GetMobileMoneyTransactions(r.Context(), userID, page, pageSize)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.APIError{Error: "failed to fetch mobile money transactions"})
		return
	}

	respondJSON(w, http.StatusOK, models.PaginatedResponse{
		Data:       txs,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: store.Paginate(totalCount, pageSize),
	})
}

// --- Budget Goals Handler ---

type BudgetGoalHandler struct {
	store    *store.Store
	validate *validator.Validate
}

func NewBudgetGoalHandler(s *store.Store) *BudgetGoalHandler {
	return &BudgetGoalHandler{
		store:    s,
		validate: validator.New(),
	}
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
		respondJSON(w, http.StatusNotFound, models.APIError{Error: "budget goal not found"})
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
		respondJSON(w, http.StatusNotFound, models.APIError{Error: "budget goal not found"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Message: "budget goal deleted"})
}

// --- Bills Handler ---

type BillHandler struct {
	store    *store.Store
	validate *validator.Validate
}

func NewBillHandler(s *store.Store) *BillHandler {
	return &BillHandler{
		store:    s,
		validate: validator.New(),
	}
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
		respondJSON(w, http.StatusNotFound, models.APIError{Error: "bill not found"})
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
		respondJSON(w, http.StatusNotFound, models.APIError{Error: "bill not found"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Message: "bill deleted"})
}

// --- Reports Handler ---

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

	// Parse date range
	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	var err error

	if startStr != "" {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid start_date format, use YYYY-MM-DD"})
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0) // Default: last month
	}

	if endStr != "" {
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, models.APIError{Error: "invalid end_date format, use YYYY-MM-DD"})
			return
		}
	} else {
		endDate = time.Now()
	}

	report, err := h.store.GetExpenditureReport(r.Context(), userID, startDate, endDate)
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
		"total_income":  totalIncome,
		"total_expense": totalExpense,
		"net_balance":   totalIncome - totalExpense,
	}})
}

// --- Budget Templates Handler ---

type BudgetTemplateHandler struct {
	store    *store.Store
	validate *validator.Validate
}

func NewBudgetTemplateHandler(s *store.Store) *BudgetTemplateHandler {
	return &BudgetTemplateHandler{
		store:    s,
		validate: validator.New(),
	}
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
		respondJSON(w, http.StatusNotFound, models.APIError{Error: "template not found"})
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{Message: "template deleted"})
}

// --- Helper ---

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
