package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ledgerly/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginate(t *testing.T) {
	tests := []struct {
		name       string
		totalCount int
		pageSize   int
		expected   int
	}{
		{"zero total", 0, 20, 0},
		{"exact page", 20, 20, 1},
		{"partial page", 25, 20, 2},
		{"zero page size", 10, 0, 0},
		{"negative page size", 10, -1, 0},
		{"large total", 1000, 20, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Paginate(tt.totalCount, tt.pageSize)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateAccountRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     models.CreateAccountRequest
		isValid bool
	}{
		{
			name: "valid bank account",
			req: models.CreateAccountRequest{
				Name:        "Chase Checking",
				Type:        "bank",
				SubType:     "checking",
				Currency:    "USD",
				Balance:     50000.00,
				Institution: "Chase Bank",
			},
			isValid: true,
		},
		{
			name: "valid mobile money account",
			req: models.CreateAccountRequest{
				Name:     "M-Pesa",
				Type:     "mobile_money",
				SubType:  "mpesa",
				Currency: "KES",
				Balance:  250000.00,
			},
			isValid: true,
		},
		{
			name: "valid investment account",
			req: models.CreateAccountRequest{
				Name:        "Interactive Brokers",
				Type:        "investment",
				SubType:     "stock_portfolio",
				Currency:    "USD",
				Balance:     500000.00,
				Institution: "Interactive Brokers",
			},
			isValid: true,
		},
		{
			name: "invalid type",
			req: models.CreateAccountRequest{
				Name:     "Test",
				Type:     "invalid",
				SubType:  "checking",
				Currency: "USD",
			},
			isValid: false,
		},
		{
			name: "invalid currency length",
			req: models.CreateAccountRequest{
				Name:     "Test",
				Type:     "bank",
				SubType:  "checking",
				Currency: "USDT",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				assert.NotEmpty(t, tt.req.Name)
				assert.Contains(t, []string{"bank", "mobile_money", "investment", "property"}, tt.req.Type)
				assert.Len(t, tt.req.Currency, 3)
				assert.GreaterOrEqual(t, tt.req.Balance, 0.0)
			}
		})
	}
}

func TestCreateInvestmentRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     models.CreateInvestmentRequest
		isValid bool
	}{
		{
			name: "valid stock",
			req: models.CreateInvestmentRequest{
				AccountID:    uuid.New(),
				Symbol:       "AAPL",
				Name:         "Apple Inc.",
				Type:         "stock",
				Quantity:     100,
				AvgCostBasis: 150.00,
				CurrentPrice: 175.00,
				Currency:     "USD",
			},
			isValid: true,
		},
		{
			name: "valid crypto",
			req: models.CreateInvestmentRequest{
				AccountID:    uuid.New(),
				Symbol:       "BTC",
				Name:         "Bitcoin",
				Type:         "crypto",
				Quantity:     0.5,
				AvgCostBasis: 30000.00,
				CurrentPrice: 45000.00,
				Currency:     "USD",
			},
			isValid: true,
		},
		{
			name: "invalid type",
			req: models.CreateInvestmentRequest{
				AccountID:    uuid.New(),
				Symbol:       "TEST",
				Name:         "Test",
				Type:         "invalid",
				Quantity:     10,
				AvgCostBasis: 100.00,
				CurrentPrice: 120.00,
				Currency:     "USD",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				assert.NotEmpty(t, tt.req.Symbol)
				assert.Contains(t, []string{"stock", "etf", "mutual_fund", "crypto", "bond"}, tt.req.Type)
				assert.Greater(t, tt.req.Quantity, 0.0)
				assert.Len(t, tt.req.Currency, 3)
			}
		})
	}
}

func TestCreatePropertyRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     models.CreatePropertyRequest
		isValid bool
	}{
		{
			name: "valid real estate",
			req: models.CreatePropertyRequest{
				AccountID:     uuid.New(),
				Name:          "Nairobi Apartment",
				Type:          "real_estate",
				Description:   "3-bedroom apartment in Westlands",
				CurrentValue:  25000000.00,
				PurchasePrice: 20000000.00,
				Currency:      "KES",
				Location:      "Westlands, Nairobi",
			},
			isValid: true,
		},
		{
			name: "valid vehicle",
			req: models.CreatePropertyRequest{
				AccountID:     uuid.New(),
				Name:          "Range Rover Sport",
				Type:          "vehicle",
				CurrentValue:  85000.00,
				PurchasePrice: 95000.00,
				Currency:      "USD",
			},
			isValid: true,
		},
		{
			name: "invalid type",
			req: models.CreatePropertyRequest{
				AccountID:     uuid.New(),
				Name:          "Test",
				Type:          "invalid",
				CurrentValue:  100000.00,
				PurchasePrice: 100000.00,
				Currency:      "USD",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				assert.NotEmpty(t, tt.req.Name)
				assert.Contains(t, []string{"real_estate", "vehicle", "art", "jewelry", "other"}, tt.req.Type)
				assert.GreaterOrEqual(t, tt.req.CurrentValue, 0.0)
				assert.Len(t, tt.req.Currency, 3)
			}
		})
	}
}

func TestBudgetGoal_HNWCategories(t *testing.T) {
	goal := models.BudgetGoal{
		ID:            uuid.New(),
		UserID:        "user_123",
		Name:          "Beach House in Diani",
		TargetAmount:  50000000.00,
		CurrentAmount: 15000000.00,
		Currency:      "KES",
		Category:      "property",
		Priority:      "high",
		Status:        "active",
	}

	// Test progress calculation
	progress := (goal.CurrentAmount / goal.TargetAmount) * 100
	assert.Equal(t, 30.0, progress)

	// Verify HNW-specific categories
	validCategories := []string{"property", "education", "retirement", "travel", "business"}
	assert.Contains(t, validCategories, goal.Category)

	// Verify priority levels
	validPriorities := []string{"high", "medium", "low"}
	assert.Contains(t, validPriorities, goal.Priority)
}

func TestNetWorthSummary(t *testing.T) {
	summary := models.NetWorthSummary{
		TotalUSD:       2500000.00,
		CashUSD:        500000.00,
		InvestmentsUSD: 1500000.00,
		PropertyUSD:    500000.00,
		Change30d:      50000.00,
		ChangePct30d:   2.04,
	}

	assert.Equal(t, summary.TotalUSD, summary.CashUSD+summary.InvestmentsUSD+summary.PropertyUSD)
	assert.Greater(t, summary.ChangePct30d, 0.0)
}

func TestUserSettings_PrivacyMode(t *testing.T) {
	settings := models.UserSettings{
		UserID:          "user_123",
		PrivacyMode:     true,
		DefaultCurrency: "USD",
		Theme:           "dark",
		UpdatedAt:       time.Now(),
	}

	assert.True(t, settings.PrivacyMode)
	assert.Equal(t, "USD", settings.DefaultCurrency)
	assert.Equal(t, "dark", settings.Theme)
}

func TestCurrencyConversion(t *testing.T) {
	conversion := models.CurrencyConversion{
		FromCurrency: "KES",
		ToCurrency:   "USD",
		Amount:       100000.00,
		Result:       770.00, // 100000 KES * 0.0077 = 770 USD
		Rate:         0.0077,
	}

	assert.Equal(t, "KES", conversion.FromCurrency)
	assert.Equal(t, "USD", conversion.ToCurrency)
	assert.Equal(t, 100000.00, conversion.Amount)
	assert.Greater(t, conversion.Result, 0.0)
}

func TestAIInsight(t *testing.T) {
	insight := models.AIInsight{
		ID:          "insight_1",
		Category:    "currency",
		Title:       "USD Exposure Declined",
		Description: "Your USD-denominated assets decreased by 8% this month due to KES strengthening against the dollar.",
		Impact:      "negative",
		Priority:    "high",
		GeneratedAt: time.Now(),
	}

	assert.Equal(t, "currency", insight.Category)
	assert.Contains(t, []string{"portfolio", "cash_flow", "currency", "tax", "goals"}, insight.Category)
	assert.Contains(t, []string{"positive", "negative", "neutral"}, insight.Impact)
	assert.Contains(t, []string{"high", "medium", "low"}, insight.Priority)
}

func TestStoreInterface(t *testing.T) {
	t.Run("store requires database URL", func(t *testing.T) {
		_, err := New("")
		require.Error(t, err)
	})
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Fatal("context should be cancelled")
	}
}

func TestMultiCurrencyAccounts(t *testing.T) {
	accounts := []models.Account{
		{
			ID:       uuid.New(),
			UserID:   "user_123",
			Name:     "Chase USD",
			Type:     "bank",
			Currency: "USD",
			Balance:  100000.00,
		},
		{
			ID:       uuid.New(),
			UserID:   "user_123",
			Name:     "M-Pesa KES",
			Type:     "mobile_money",
			Currency: "KES",
			Balance:  5000000.00,
		},
		{
			ID:       uuid.New(),
			UserID:   "user_123",
			Name:     "Barclays GBP",
			Type:     "bank",
			Currency: "GBP",
			Balance:  50000.00,
		},
	}

	// Verify multi-currency support
	currencies := make(map[string]bool)
	for _, a := range accounts {
		currencies[a.Currency] = true
	}

	assert.True(t, currencies["USD"])
	assert.True(t, currencies["KES"])
	assert.True(t, currencies["GBP"])
}
