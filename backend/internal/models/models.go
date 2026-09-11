package models

import (
	"time"

	"github.com/google/uuid"
)

// Transaction represents an income or expense entry
type Transaction struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      string     `json:"user_id" db:"user_id"`
	Amount      float64    `json:"amount" db:"amount"`
	Type        string     `json:"type" db:"type"` // "income" or "expense"
	Category    string     `json:"category" db:"category"`
	Description string     `json:"description" db:"description"`
	Source      string     `json:"source" db:"source"` // "bank", "cash", "mobile_money", "card"
	Date        time.Time  `json:"date" db:"date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// MobileMoneyTransaction represents M-Pesa style mobile money transactions
type MobileMoneyTransaction struct {
	ID                uuid.UUID `json:"id" db:"id"`
	UserID            string    `json:"user_id" db:"user_id"`
	TransactionID     string    `json:"transaction_id" db:"transaction_id"` // External reference ID
	Amount            float64   `json:"amount" db:"amount"`
	Type              string    `json:"type" db:"type"` // "send", "receive", "paybill", "buygoods"
	PhoneNumber       string    `json:"phone_number" db:"phone_number"`
	Provider          string    `json:"provider" db:"provider"` // "mpesa", "airtel_money", "tigo_pesa"
	CounterpartyName  string    `json:"counterparty_name" db:"counterparty_name"`
	Description       string    `json:"description" db:"description"`
	Status            string    `json:"status" db:"status"` // "completed", "pending", "failed"
	Date              time.Time `json:"date" db:"date"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// BudgetGoal represents a savings goal with target and progress
type BudgetGoal struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	TargetAmount float64  `json:"target_amount" db:"target_amount"`
	CurrentAmount float64 `json:"current_amount" db:"current_amount"`
	Category    string    `json:"category" db:"category"`
	Deadline    *time.Time `json:"deadline,omitempty" db:"deadline"`
	Status      string    `json:"status" db:"status"` // "active", "completed", "paused"
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Bill represents a recurring bill
type Bill struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      string     `json:"user_id" db:"user_id"`
	Name        string     `json:"name" db:"name"`
	Amount      float64    `json:"amount" db:"amount"`
	Category    string     `json:"category" db:"category"`
	DueDate     time.Time  `json:"due_date" db:"due_date"`
	Recurrence  string     `json:"recurrence" db:"recurrence"` // "monthly", "weekly", "yearly", "once"
	IsPaid      bool       `json:"is_paid" db:"is_paid"`
	PaidDate    *time.Time `json:"paid_date,omitempty" db:"paid_date"`
	Notes       string     `json:"notes" db:"notes"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// BudgetTemplate represents a reusable budget structure
type BudgetTemplate struct {
	ID          uuid.UUID          `json:"id" db:"id"`
	UserID      string             `json:"user_id" db:"user_id"`
	Name        string             `json:"name" db:"name"`
	Description string             `json:"description" db:"description"`
	Categories  []TemplateCategory `json:"categories"`
	IsPublic    bool               `json:"is_public" db:"is_public"`
	CreatedAt   time.Time          `json:"created_at" db:"created_at"`
}

// TemplateCategory is a category allocation within a budget template
type TemplateCategory struct {
	ID           uuid.UUID `json:"id" db:"id"`
	TemplateID   uuid.UUID `json:"template_id" db:"template_id"`
	Category     string    `json:"category" db:"category"`
	AllocatedPct float64   `json:"allocated_pct" db:"allocated_pct"` // percentage of income
	LimitAmount  *float64  `json:"limit_amount,omitempty" db:"limit_amount"`
}

// ExpenditureReport represents spending breakdown
type ExpenditureReport struct {
	StartDate  time.Time          `json:"start_date"`
	EndDate    time.Time          `json:"end_date"`
	Categories []CategorySpending `json:"categories"`
	TotalSpent float64            `json:"total_spent"`
}

// CategorySpending is spending in a single category
type CategorySpending struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Count    int     `json:"count"`
}

// ChatMessage for AI assistant
type ChatMessage struct {
	Role    string `json:"role" validate:"required,oneof=user assistant"`
	Content string `json:"content" validate:"required"`
}

// ChatRequest for AI chat endpoint
type ChatRequest struct {
	Messages []ChatMessage `json:"messages" validate:"required,min=1"`
}

// BudgetSuggestion from AI
type BudgetSuggestion struct {
	Categories []SuggestedCategory `json:"categories"`
	TotalIncome float64            `json:"total_income"`
	Notes      string              `json:"notes"`
}

// SuggestedCategory in a budget suggestion
type SuggestedCategory struct {
	Category     string  `json:"category"`
	Recommended  float64 `json:"recommended"`
	Percentage   float64 `json:"percentage"`
	Rationale    string  `json:"rationale"`
}

// CreateTransactionRequest
type CreateTransactionRequest struct {
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Type        string    `json:"type" validate:"required,oneof=income expense"`
	Category    string    `json:"category" validate:"required"`
	Description string    `json:"description"`
	Source      string    `json:"source" validate:"required,oneof=bank cash mobile_money card"`
	Date        time.Time `json:"date"`
}

// CreateMobileMoneyRequest
type CreateMobileMoneyRequest struct {
	TransactionID    string  `json:"transaction_id" validate:"required"`
	Amount           float64 `json:"amount" validate:"required,gt=0"`
	Type             string  `json:"type" validate:"required,oneof=send receive paybill buygoods"`
	PhoneNumber      string  `json:"phone_number" validate:"required"`
	Provider         string  `json:"provider" validate:"required,oneof=mpesa airtel_money tigo_pesa"`
	CounterpartyName string  `json:"counterparty_name"`
	Description      string  `json:"description"`
	Date             time.Time `json:"date"`
}

// CreateBudgetGoalRequest
type CreateBudgetGoalRequest struct {
	Name         string     `json:"name" validate:"required"`
	TargetAmount float64    `json:"target_amount" validate:"required,gt=0"`
	Category     string     `json:"category" validate:"required"`
	Deadline     *time.Time `json:"deadline,omitempty"`
}

// CreateBillRequest
type CreateBillRequest struct {
	Name       string    `json:"name" validate:"required"`
	Amount     float64   `json:"amount" validate:"required,gt=0"`
	Category   string    `json:"category" validate:"required"`
	DueDate    time.Time `json:"due_date" validate:"required"`
	Recurrence string    `json:"recurrence" validate:"required,oneof=monthly weekly yearly once"`
	Notes      string    `json:"notes"`
}

// CreateBudgetTemplateRequest
type CreateBudgetTemplateRequest struct {
	Name        string             `json:"name" validate:"required"`
	Description string             `json:"description"`
	Categories  []TemplateCategory `json:"categories" validate:"required,min=1"`
	IsPublic    bool               `json:"is_public"`
}

// APIError represents a standard error response
type APIError struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// APIResponse is a standard success response
type APIResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PaginationParams for list endpoints
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// PaginatedResponse wraps paginated results
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int         `json:"total_count"`
	TotalPages int         `json:"total_pages"`
}
