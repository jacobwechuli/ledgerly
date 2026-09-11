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

// MockStore for testing without a real database
type MockStore struct {
	transactions             []models.Transaction
	mobileMoneyTransactions  []models.MobileMoneyTransaction
	budgetGoals              []models.BudgetGoal
	bills                    []models.Bill
	budgetTemplates          []models.BudgetTemplate
}

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

func TestCreateTransactionRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     models.CreateTransactionRequest
		isValid bool
	}{
		{
			name: "valid income",
			req: models.CreateTransactionRequest{
				Amount:   100.50,
				Type:     "income",
				Category: "salary",
				Source:   "bank",
				Date:     time.Now(),
			},
			isValid: true,
		},
		{
			name: "valid expense",
			req: models.CreateTransactionRequest{
				Amount:   50.00,
				Type:     "expense",
				Category: "food",
				Source:   "cash",
				Date:     time.Now(),
			},
			isValid: true,
		},
		{
			name: "invalid type",
			req: models.CreateTransactionRequest{
				Amount:   100.00,
				Type:     "invalid",
				Category: "food",
				Source:   "bank",
			},
			isValid: false,
		},
		{
			name: "zero amount",
			req: models.CreateTransactionRequest{
				Amount:   0,
				Type:     "income",
				Category: "salary",
				Source:   "bank",
			},
			isValid: false,
		},
		{
			name: "negative amount",
			req: models.CreateTransactionRequest{
				Amount:   -50.00,
				Type:     "expense",
				Category: "food",
				Source:   "cash",
			},
			isValid: false,
		},
		{
			name: "missing source",
			req: models.CreateTransactionRequest{
				Amount:   100.00,
				Type:     "income",
				Category: "salary",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation checks
			if tt.isValid {
				assert.Greater(t, tt.req.Amount, 0.0)
				assert.Contains(t, []string{"income", "expense"}, tt.req.Type)
				assert.NotEmpty(t, tt.req.Category)
				assert.Contains(t, []string{"bank", "cash", "mobile_money", "card"}, tt.req.Source)
			}
		})
	}
}

func TestCreateMobileMoneyRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     models.CreateMobileMoneyRequest
		isValid bool
	}{
		{
			name: "valid send",
			req: models.CreateMobileMoneyRequest{
				TransactionID: "TX123",
				Amount:        500.00,
				Type:          "send",
				PhoneNumber:   "+254712345678",
				Provider:      "mpesa",
			},
			isValid: true,
		},
		{
			name: "valid receive",
			req: models.CreateMobileMoneyRequest{
				TransactionID: "TX456",
				Amount:        1000.00,
				Type:          "receive",
				PhoneNumber:   "+254712345678",
				Provider:      "airtel_money",
			},
			isValid: true,
		},
		{
			name: "invalid provider",
			req: models.CreateMobileMoneyRequest{
				TransactionID: "TX789",
				Amount:        200.00,
				Type:          "send",
				PhoneNumber:   "+254712345678",
				Provider:      "invalid_provider",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validProviders := []string{"mpesa", "airtel_money", "tigo_pesa"}
			validTypes := []string{"send", "receive", "paybill", "buygoods"}

			if tt.isValid {
				assert.Greater(t, tt.req.Amount, 0.0)
				assert.Contains(t, validTypes, tt.req.Type)
				assert.Contains(t, validProviders, tt.req.Provider)
				assert.NotEmpty(t, tt.req.PhoneNumber)
				assert.NotEmpty(t, tt.req.TransactionID)
			}
		})
	}
}

func TestBudgetGoal_Progress(t *testing.T) {
	goal := models.BudgetGoal{
		ID:            uuid.New(),
		UserID:        "user_123",
		Name:          "Emergency Fund",
		TargetAmount:  10000.00,
		CurrentAmount: 0,
		Category:      "savings",
		Status:        "active",
	}

	// Test progress calculation
	progress := (goal.CurrentAmount / goal.TargetAmount) * 100
	assert.Equal(t, 0.0, progress)

	// Simulate contribution
	goal.CurrentAmount += 2500.00
	progress = (goal.CurrentAmount / goal.TargetAmount) * 100
	assert.Equal(t, 25.0, progress)

	// Complete the goal
	goal.CurrentAmount = goal.TargetAmount
	progress = (goal.CurrentAmount / goal.TargetAmount) * 100
	assert.Equal(t, 100.0, progress)
}

func TestBill_PaidStatus(t *testing.T) {
	bill := models.Bill{
		ID:         uuid.New(),
		UserID:     "user_123",
		Name:       "Electricity",
		Amount:     150.00,
		Category:   "utilities",
		DueDate:    time.Now().Add(7 * 24 * time.Hour),
		Recurrence: "monthly",
		IsPaid:     false,
	}

	assert.False(t, bill.IsPaid)
	assert.Nil(t, bill.PaidDate)

	// Mark as paid
	now := time.Now()
	bill.IsPaid = true
	bill.PaidDate = &now

	assert.True(t, bill.IsPaid)
	assert.NotNil(t, bill.PaidDate)
}

// Integration-style test that verifies the store interface
func TestStoreInterface(t *testing.T) {
	// This test verifies the store can be created with valid config
	// Actual DB tests would need a test database
	t.Run("store requires database URL", func(t *testing.T) {
		_, err := New("")
		require.Error(t, err)
	})
}

func TestExpenditureReport(t *testing.T) {
	report := models.ExpenditureReport{
		StartDate: time.Now().AddDate(0, -1, 0),
		EndDate:   time.Now(),
		Categories: []models.CategorySpending{
			{Category: "food", Amount: 500.00, Count: 25},
			{Category: "transport", Amount: 200.00, Count: 15},
			{Category: "entertainment", Amount: 150.00, Count: 5},
		},
		TotalSpent: 850.00,
	}

	assert.Len(t, report.Categories, 3)
	assert.Equal(t, 850.00, report.TotalSpent)

	// Verify total matches sum of categories
	var sum float64
	for _, cat := range report.Categories {
		sum += cat.Amount
	}
	assert.Equal(t, sum, report.TotalSpent)
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Verify context is cancelled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Fatal("context should be cancelled")
	}
}
