package models

import (
	"time"

	"github.com/google/uuid"
)

// Account represents a financial account (bank, mobile money, investment, property)
type Account struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Type        string    `json:"type" db:"type"` // "bank", "mobile_money", "investment", "property"
	SubType     string    `json:"sub_type" db:"sub_type"` // "checking", "savings", "mpesa", "stock_portfolio", "crypto_wallet", "real_estate", "vehicle"
	Currency    string    `json:"currency" db:"currency"` // ISO 4217: "USD", "KES", "GBP", etc.
	Balance     float64   `json:"balance" db:"balance"`
	Institution string    `json:"institution,omitempty" db:"institution"` // Bank name, broker, etc.
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Transaction represents an income or expense entry
type Transaction struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	AccountID   uuid.UUID `json:"account_id" db:"account_id"`
	Amount      float64   `json:"amount" db:"amount"`
	Currency    string    `json:"currency" db:"currency"` // Transaction currency
	Type        string    `json:"type" db:"type"`         // "income", "expense", "transfer"
	Category    string    `json:"category" db:"category"`
	Description string    `json:"description" db:"description"`
	Date        time.Time `json:"date" db:"date"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Investment represents a holding in stocks, funds, or crypto
type Investment struct {
	ID            uuid.UUID `json:"id" db:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	AccountID     uuid.UUID `json:"account_id" db:"account_id"`
	Symbol        string    `json:"symbol" db:"symbol"` // Ticker: AAPL, BTC, etc.
	Name          string    `json:"name" db:"name"`
	Type          string    `json:"type" db:"type"` // "stock", "etf", "mutual_fund", "crypto", "bond"
	Quantity      float64   `json:"quantity" db:"quantity"`
	AvgCostBasis  float64   `json:"avg_cost_basis" db:"avg_cost_basis"`
	CurrentPrice  float64   `json:"current_price" db:"current_price"`
	Currency      string    `json:"currency" db:"currency"`
	UnrealizedGain float64  `json:"unrealized_gain" db:"unrealized_gain"`
	LastUpdated   time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// Property represents real estate or vehicle assets
type Property struct {
	ID            uuid.UUID `json:"id" db:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	AccountID     uuid.UUID `json:"account_id" db:"account_id"`
	Name          string    `json:"name" db:"name"`
	Type          string    `json:"type" db:"type"` // "real_estate", "vehicle", "art", "jewelry", "other"
	Description   string    `json:"description" db:"description"`
	CurrentValue  float64   `json:"current_value" db:"current_value"`
	PurchasePrice float64   `json:"purchase_price" db:"purchase_price"`
	Currency      string    `json:"currency" db:"currency"`
	Location      string    `json:"location,omitempty" db:"location"`
	PurchaseDate  time.Time `json:"purchase_date" db:"purchase_date"`
	LastValued    time.Time `json:"last_valued" db:"last_valued"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// NetWorthSnapshot represents net worth at a point in time
type NetWorthSnapshot struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Date      time.Time `json:"date" db:"date"`
	TotalUSD  float64   `json:"total_usd" db:"total_usd"` // Aggregated in USD
	CashUSD   float64   `json:"cash_usd" db:"cash_usd"`
	InvestUSD float64   `json:"invest_usd" db:"invest_usd"`
	PropUSD   float64   `json:"prop_usd" db:"prop_usd"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Currency represents exchange rate data
type Currency struct {
	Code      string    `json:"code" db:"code"` // ISO 4217
	Name      string    `json:"name" db:"name"`
	Symbol    string    `json:"symbol" db:"symbol"`
	RateToUSD float64   `json:"rate_to_usd" db:"rate_to_usd"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Bill represents a recurring bill
type Bill struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	UserID     string     `json:"user_id" db:"user_id"`
	AccountID  uuid.UUID  `json:"account_id" db:"account_id"`
	Name       string     `json:"name" db:"name"`
	Amount     float64    `json:"amount" db:"amount"`
	Currency   string     `json:"currency" db:"currency"`
	Category   string     `json:"category" db:"category"`
	DueDate    time.Time  `json:"due_date" db:"due_date"`
	Recurrence string     `json:"recurrence" db:"recurrence"` // "monthly", "weekly", "yearly", "once"
	IsPaid     bool       `json:"is_paid" db:"is_paid"`
	PaidDate   *time.Time `json:"paid_date,omitempty" db:"paid_date"`
	Notes      string     `json:"notes" db:"notes"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// BudgetGoal represents a large financial goal (property, school fees, etc.)
type BudgetGoal struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        string     `json:"user_id" db:"user_id"`
	Name          string     `json:"name" db:"name"`
	TargetAmount  float64    `json:"target_amount" db:"target_amount"`
	CurrentAmount float64    `json:"current_amount" db:"current_amount"`
	Currency      string     `json:"currency" db:"currency"`
	Category      string     `json:"category" db:"category"` // "property", "education", "retirement", "travel", "business"
	Deadline      *time.Time `json:"deadline,omitempty" db:"deadline"`
	Status        string     `json:"status" db:"status"` // "active", "completed", "paused"
	Priority      string     `json:"priority" db:"priority"` // "high", "medium", "low"
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
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
	AllocatedPct float64   `json:"allocated_pct" db:"allocated_pct"`
	LimitAmount  *float64  `json:"limit_amount,omitempty" db:"limit_amount"`
}

// UserSettings represents user preferences
type UserSettings struct {
	UserID       string `json:"user_id" db:"user_id"`
	PrivacyMode  bool   `json:"privacy_mode" db:"privacy_mode"` // Mask balances by default
	DefaultCurrency string `json:"default_currency" db:"default_currency"`
	Theme        string `json:"theme" db:"theme"` // "light", "dark", "system"
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// TaxRecord represents tax liability data
type TaxRecord struct {
	ID            uuid.UUID `json:"id" db:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	Year          int       `json:"year" db:"year"`
	Type          string    `json:"type" db:"type"` // "capital_gains", "income_tax", "dividend_tax"
	Amount        float64   `json:"amount" db:"amount"`
	Currency      string    `json:"currency" db:"currency"`
	Description   string    `json:"description" db:"description"`
	EstimatedAt   time.Time `json:"estimated_at" db:"estimated_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// ExpenditureReport represents spending breakdown
type ExpenditureReport struct {
	StartDate  time.Time          `json:"start_date"`
	EndDate    time.Time          `json:"end_date"`
	Categories []CategorySpending `json:"categories"`
	TotalSpent float64            `json:"total_spent"`
	Currency   string             `json:"currency"`
}

// CategorySpending is spending in a single category
type CategorySpending struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Count    int     `json:"count"`
}

// NetWorthSummary represents current net worth breakdown
type NetWorthSummary struct {
	TotalUSD      float64 `json:"total_usd"`
	CashUSD       float64 `json:"cash_usd"`
	InvestmentsUSD float64 `json:"investments_usd"`
	PropertyUSD   float64 `json:"property_usd"`
	Change24h     float64 `json:"change_24h"`
	ChangePct24h  float64 `json:"change_pct_24h"`
	Change30d     float64 `json:"change_30d"`
	ChangePct30d  float64 `json:"change_pct_30d"`
}

// CurrencyConversion represents a currency conversion result
type CurrencyConversion struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	Amount       float64 `json:"amount"`
	Result       float64 `json:"result"`
	Rate         float64 `json:"rate"`
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

// AIInsight represents a concierge-style observation
type AIInsight struct {
	ID          string    `json:"id"`
	Category    string    `json:"category"` // "portfolio", "cash_flow", "currency", "tax", "goals"
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"` // "positive", "negative", "neutral"
	Priority    string    `json:"priority"` // "high", "medium", "low"
	GeneratedAt time.Time `json:"generated_at"`
}

// BudgetSuggestion from AI
type BudgetSuggestion struct {
	Categories  []SuggestedCategory `json:"categories"`
	TotalIncome float64             `json:"total_income"`
	Notes       string              `json:"notes"`
}

// SuggestedCategory in a budget suggestion
type SuggestedCategory struct {
	Category   string  `json:"category"`
	Recommended float64 `json:"recommended"`
	Percentage float64 `json:"percentage"`
	Rationale  string  `json:"rationale"`
}

// Request types

// CreateAccountRequest
type CreateAccountRequest struct {
	Name        string  `json:"name" validate:"required"`
	Type        string  `json:"type" validate:"required,oneof=bank mobile_money investment property"`
	SubType     string  `json:"sub_type" validate:"required"`
	Currency    string  `json:"currency" validate:"required,len=3"`
	Balance     float64 `json:"balance" validate:"gte=0"`
	Institution string  `json:"institution"`
}

// CreateTransactionRequest
type CreateTransactionRequest struct {
	AccountID   uuid.UUID `json:"account_id" validate:"required"`
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Currency    string    `json:"currency" validate:"required,len=3"`
	Type        string    `json:"type" validate:"required,oneof=income expense transfer"`
	Category    string    `json:"category" validate:"required"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

// CreateInvestmentRequest
type CreateInvestmentRequest struct {
	AccountID    uuid.UUID `json:"account_id" validate:"required"`
	Symbol       string    `json:"symbol" validate:"required"`
	Name         string    `json:"name" validate:"required"`
	Type         string    `json:"type" validate:"required,oneof=stock etf mutual_fund crypto bond"`
	Quantity     float64   `json:"quantity" validate:"required,gt=0"`
	AvgCostBasis float64   `json:"avg_cost_basis" validate:"required,gte=0"`
	CurrentPrice float64   `json:"current_price" validate:"required,gte=0"`
	Currency     string    `json:"currency" validate:"required,len=3"`
}

// CreatePropertyRequest
type CreatePropertyRequest struct {
	AccountID     uuid.UUID `json:"account_id" validate:"required"`
	Name          string    `json:"name" validate:"required"`
	Type          string    `json:"type" validate:"required,oneof=real_estate vehicle art jewelry other"`
	Description   string    `json:"description"`
	CurrentValue  float64   `json:"current_value" validate:"required,gte=0"`
	PurchasePrice float64   `json:"purchase_price" validate:"required,gte=0"`
	Currency      string    `json:"currency" validate:"required,len=3"`
	Location      string    `json:"location"`
	PurchaseDate  time.Time `json:"purchase_date"`
}

// CreateBillRequest
type CreateBillRequest struct {
	AccountID  uuid.UUID `json:"account_id" validate:"required"`
	Name       string    `json:"name" validate:"required"`
	Amount     float64   `json:"amount" validate:"required,gt=0"`
	Currency   string    `json:"currency" validate:"required,len=3"`
	Category   string    `json:"category" validate:"required"`
	DueDate    time.Time `json:"due_date" validate:"required"`
	Recurrence string    `json:"recurrence" validate:"required,oneof=monthly weekly yearly once"`
	Notes      string    `json:"notes"`
}

// CreateBudgetGoalRequest
type CreateBudgetGoalRequest struct {
	Name         string     `json:"name" validate:"required"`
	TargetAmount float64    `json:"target_amount" validate:"required,gt=0"`
	Currency     string     `json:"currency" validate:"required,len=3"`
	Category     string     `json:"category" validate:"required,oneof=property education retirement travel business"`
	Deadline     *time.Time `json:"deadline,omitempty"`
	Priority     string     `json:"priority" validate:"required,oneof=high medium low"`
}

// CreateBudgetTemplateRequest
type CreateBudgetTemplateRequest struct {
	Name        string             `json:"name" validate:"required"`
	Description string             `json:"description"`
	Categories  []TemplateCategory `json:"categories" validate:"required,min=1"`
	IsPublic    bool               `json:"is_public"`
}

// UpdateUserSettingsRequest
type UpdateUserSettingsRequest struct {
	PrivacyMode     *bool   `json:"privacy_mode,omitempty"`
	DefaultCurrency *string `json:"default_currency,omitempty"`
	Theme           *string `json:"theme,omitempty"`
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

// PaginatedResponse wraps paginated results
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int         `json:"total_count"`
	TotalPages int         `json:"total_pages"`
}
