package store

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ledgerly/backend/internal/models"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(databaseURL string) (*Store, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	s.pool.Close()
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

// --- Transactions ---

func (s *Store) CreateTransaction(ctx context.Context, userID string, req models.CreateTransactionRequest) (*models.Transaction, error) {
	tx := &models.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		Amount:      req.Amount,
		Type:        req.Type,
		Category:    req.Category,
		Description: req.Description,
		Source:      req.Source,
		Date:        req.Date,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if tx.Date.IsZero() {
		tx.Date = time.Now()
	}

	query := `
		INSERT INTO transactions (id, user_id, amount, type, category, description, source, date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, user_id, amount, type, category, description, source, date, created_at, updated_at
	`

	err := s.pool.QueryRow(ctx, query,
		tx.ID, tx.UserID, tx.Amount, tx.Type, tx.Category,
		tx.Description, tx.Source, tx.Date, tx.CreatedAt, tx.UpdatedAt,
	).Scan(&tx.ID, &tx.UserID, &tx.Amount, &tx.Type, &tx.Category,
		&tx.Description, &tx.Source, &tx.Date, &tx.CreatedAt, &tx.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return tx, nil
}

func (s *Store) GetTransactions(ctx context.Context, userID string, page, pageSize int) ([]models.Transaction, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	var totalCount int
	countQuery := `SELECT COUNT(*) FROM transactions WHERE user_id = $1`
	err := s.pool.QueryRow(ctx, countQuery, userID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	query := `
		SELECT id, user_id, amount, type, category, description, source, date, created_at, updated_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.Type, &t.Category,
			&t.Description, &t.Source, &t.Date, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, t)
	}

	return transactions, totalCount, nil
}

func (s *Store) GetTransactionByID(ctx context.Context, userID string, id uuid.UUID) (*models.Transaction, error) {
	query := `
		SELECT id, user_id, amount, type, category, description, source, date, created_at, updated_at
		FROM transactions
		WHERE id = $1 AND user_id = $2
	`

	var t models.Transaction
	err := s.pool.QueryRow(ctx, query, id, userID).Scan(
		&t.ID, &t.UserID, &t.Amount, &t.Type, &t.Category,
		&t.Description, &t.Source, &t.Date, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	return &t, nil
}

func (s *Store) DeleteTransaction(ctx context.Context, userID string, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx,
		`DELETE FROM transactions WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("transaction not found")
	}
	return nil
}

// --- Mobile Money Transactions ---

func (s *Store) CreateMobileMoneyTransaction(ctx context.Context, userID string, req models.CreateMobileMoneyRequest) (*models.MobileMoneyTransaction, error) {
	tx := &models.MobileMoneyTransaction{
		ID:               uuid.New(),
		UserID:           userID,
		TransactionID:    req.TransactionID,
		Amount:           req.Amount,
		Type:             req.Type,
		PhoneNumber:      req.PhoneNumber,
		Provider:         req.Provider,
		CounterpartyName: req.CounterpartyName,
		Description:      req.Description,
		Status:           "completed",
		Date:             req.Date,
		CreatedAt:        time.Now(),
	}

	if tx.Date.IsZero() {
		tx.Date = time.Now()
	}

	query := `
		INSERT INTO mobile_money_transactions (id, user_id, transaction_id, amount, type, phone_number, provider, counterparty_name, description, status, date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, user_id, transaction_id, amount, type, phone_number, provider, counterparty_name, description, status, date, created_at
	`

	err := s.pool.QueryRow(ctx, query,
		tx.ID, tx.UserID, tx.TransactionID, tx.Amount, tx.Type,
		tx.PhoneNumber, tx.Provider, tx.CounterpartyName, tx.Description,
		tx.Status, tx.Date, tx.CreatedAt,
	).Scan(&tx.ID, &tx.UserID, &tx.TransactionID, &tx.Amount, &tx.Type,
		&tx.PhoneNumber, &tx.Provider, &tx.CounterpartyName, &tx.Description,
		&tx.Status, &tx.Date, &tx.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create mobile money transaction: %w", err)
	}

	return tx, nil
}

func (s *Store) GetMobileMoneyTransactions(ctx context.Context, userID string, page, pageSize int) ([]models.MobileMoneyTransaction, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var totalCount int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM mobile_money_transactions WHERE user_id = $1`, userID).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, transaction_id, amount, type, phone_number, provider, counterparty_name, description, status, date, created_at
		FROM mobile_money_transactions
		WHERE user_id = $1
		ORDER BY date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var txs []models.MobileMoneyTransaction
	for rows.Next() {
		var t models.MobileMoneyTransaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.TransactionID, &t.Amount, &t.Type,
			&t.PhoneNumber, &t.Provider, &t.CounterpartyName, &t.Description,
			&t.Status, &t.Date, &t.CreatedAt); err != nil {
			return nil, 0, err
		}
		txs = append(txs, t)
	}

	return txs, totalCount, nil
}

// --- Budget Goals ---

func (s *Store) CreateBudgetGoal(ctx context.Context, userID string, req models.CreateBudgetGoalRequest) (*models.BudgetGoal, error) {
	goal := &models.BudgetGoal{
		ID:            uuid.New(),
		UserID:        userID,
		Name:          req.Name,
		TargetAmount:  req.TargetAmount,
		CurrentAmount: 0,
		Category:      req.Category,
		Deadline:      req.Deadline,
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	query := `
		INSERT INTO budget_goals (id, user_id, name, target_amount, current_amount, category, deadline, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, user_id, name, target_amount, current_amount, category, deadline, status, created_at, updated_at
	`

	err := s.pool.QueryRow(ctx, query,
		goal.ID, goal.UserID, goal.Name, goal.TargetAmount, goal.CurrentAmount,
		goal.Category, goal.Deadline, goal.Status, goal.CreatedAt, goal.UpdatedAt,
	).Scan(&goal.ID, &goal.UserID, &goal.Name, &goal.TargetAmount, &goal.CurrentAmount,
		&goal.Category, &goal.Deadline, &goal.Status, &goal.CreatedAt, &goal.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create budget goal: %w", err)
	}

	return goal, nil
}

func (s *Store) GetBudgetGoals(ctx context.Context, userID string) ([]models.BudgetGoal, error) {
	query := `
		SELECT id, user_id, name, target_amount, current_amount, category, deadline, status, created_at, updated_at
		FROM budget_goals
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []models.BudgetGoal
	for rows.Next() {
		var g models.BudgetGoal
		if err := rows.Scan(&g.ID, &g.UserID, &g.Name, &g.TargetAmount, &g.CurrentAmount,
			&g.Category, &g.Deadline, &g.Status, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}

	return goals, nil
}

func (s *Store) UpdateBudgetGoalProgress(ctx context.Context, userID string, goalID uuid.UUID, amount float64) (*models.BudgetGoal, error) {
	query := `
		UPDATE budget_goals
		SET current_amount = current_amount + $3, updated_at = $4,
			status = CASE WHEN current_amount + $3 >= target_amount THEN 'completed' ELSE status END
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, target_amount, current_amount, category, deadline, status, created_at, updated_at
	`

	var g models.BudgetGoal
	err := s.pool.QueryRow(ctx, query, goalID, userID, amount, time.Now()).Scan(
		&g.ID, &g.UserID, &g.Name, &g.TargetAmount, &g.CurrentAmount,
		&g.Category, &g.Deadline, &g.Status, &g.CreatedAt, &g.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("budget goal not found: %w", err)
	}

	return &g, nil
}

func (s *Store) DeleteBudgetGoal(ctx context.Context, userID string, goalID uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM budget_goals WHERE id = $1 AND user_id = $2`, goalID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("budget goal not found")
	}
	return nil
}

// --- Bills ---

func (s *Store) CreateBill(ctx context.Context, userID string, req models.CreateBillRequest) (*models.Bill, error) {
	bill := &models.Bill{
		ID:         uuid.New(),
		UserID:     userID,
		Name:       req.Name,
		Amount:     req.Amount,
		Category:   req.Category,
		DueDate:    req.DueDate,
		Recurrence: req.Recurrence,
		IsPaid:     false,
		Notes:      req.Notes,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	query := `
		INSERT INTO bills (id, user_id, name, amount, category, due_date, recurrence, is_paid, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, user_id, name, amount, category, due_date, recurrence, is_paid, paid_date, notes, created_at, updated_at
	`

	err := s.pool.QueryRow(ctx, query,
		bill.ID, bill.UserID, bill.Name, bill.Amount, bill.Category,
		bill.DueDate, bill.Recurrence, bill.IsPaid, bill.Notes,
		bill.CreatedAt, bill.UpdatedAt,
	).Scan(&bill.ID, &bill.UserID, &bill.Name, &bill.Amount, &bill.Category,
		&bill.DueDate, &bill.Recurrence, &bill.IsPaid, &bill.PaidDate,
		&bill.Notes, &bill.CreatedAt, &bill.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create bill: %w", err)
	}

	return bill, nil
}

func (s *Store) GetBills(ctx context.Context, userID string, paid *bool) ([]models.Bill, error) {
	query := `
		SELECT id, user_id, name, amount, category, due_date, recurrence, is_paid, paid_date, notes, created_at, updated_at
		FROM bills
		WHERE user_id = $1
	`
	args := []interface{}{userID}

	if paid != nil {
		query += fmt.Sprintf(" AND is_paid = $%d", len(args)+1)
		args = append(args, *paid)
	}

	query += " ORDER BY due_date ASC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bills []models.Bill
	for rows.Next() {
		var b models.Bill
		if err := rows.Scan(&b.ID, &b.UserID, &b.Name, &b.Amount, &b.Category,
			&b.DueDate, &b.Recurrence, &b.IsPaid, &b.PaidDate,
			&b.Notes, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		bills = append(bills, b)
	}

	return bills, nil
}

func (s *Store) MarkBillPaid(ctx context.Context, userID string, billID uuid.UUID) (*models.Bill, error) {
	now := time.Now()
	query := `
		UPDATE bills
		SET is_paid = true, paid_date = $3, updated_at = $3
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, amount, category, due_date, recurrence, is_paid, paid_date, notes, created_at, updated_at
	`

	var b models.Bill
	err := s.pool.QueryRow(ctx, query, billID, userID, now).Scan(
		&b.ID, &b.UserID, &b.Name, &b.Amount, &b.Category,
		&b.DueDate, &b.Recurrence, &b.IsPaid, &b.PaidDate,
		&b.Notes, &b.CreatedAt, &b.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}

	return &b, nil
}

func (s *Store) DeleteBill(ctx context.Context, userID string, billID uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM bills WHERE id = $1 AND user_id = $2`, billID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("bill not found")
	}
	return nil
}

// --- Reports ---

func (s *Store) GetExpenditureReport(ctx context.Context, userID string, startDate, endDate time.Time) (*models.ExpenditureReport, error) {
	query := `
		SELECT category, SUM(amount) as total, COUNT(*) as count
		FROM transactions
		WHERE user_id = $1 AND type = 'expense' AND date >= $2 AND date <= $3
		GROUP BY category
		ORDER BY total DESC
	`

	rows, err := s.pool.Query(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := &models.ExpenditureReport{
		StartDate: startDate,
		EndDate:   endDate,
	}

	for rows.Next() {
		var cs models.CategorySpending
		if err := rows.Scan(&cs.Category, &cs.Amount, &cs.Count); err != nil {
			return nil, err
		}
		report.Categories = append(report.Categories, cs)
		report.TotalSpent += cs.Amount
	}

	return report, nil
}

func (s *Store) GetTransactionSummary(ctx context.Context, userID string) (totalIncome, totalExpense float64, err error) {
	query := `
		SELECT type, COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1
		GROUP BY type
	`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var txType string
		var amount float64
		if err := rows.Scan(&txType, &amount); err != nil {
			return 0, 0, err
		}
		switch txType {
		case "income":
			totalIncome = amount
		case "expense":
			totalExpense = amount
		}
	}

	return totalIncome, totalExpense, nil
}

// --- Budget Templates ---

func (s *Store) CreateBudgetTemplate(ctx context.Context, userID string, req models.CreateBudgetTemplateRequest) (*models.BudgetTemplate, error) {
	template := &models.BudgetTemplate{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		Categories:  req.Categories,
		CreatedAt:   time.Now(),
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO budget_templates (id, user_id, name, description, is_public, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, template.ID, template.UserID, template.Name, template.Description, template.IsPublic, template.CreatedAt)
	if err != nil {
		return nil, err
	}

	for i := range template.Categories {
		template.Categories[i].ID = uuid.New()
		template.Categories[i].TemplateID = template.ID
		_, err = tx.Exec(ctx, `
			INSERT INTO template_categories (id, template_id, category, allocated_pct, limit_amount)
			VALUES ($1, $2, $3, $4, $5)
		`, template.Categories[i].ID, template.Categories[i].TemplateID,
			template.Categories[i].Category, template.Categories[i].AllocatedPct,
			template.Categories[i].LimitAmount)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return template, nil
}

func (s *Store) GetBudgetTemplates(ctx context.Context, userID string) ([]models.BudgetTemplate, error) {
	query := `
		SELECT id, user_id, name, description, is_public, created_at
		FROM budget_templates
		WHERE user_id = $1 OR is_public = true
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.BudgetTemplate
	for rows.Next() {
		var t models.BudgetTemplate
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Description, &t.IsPublic, &t.CreatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}

	// Load categories for each template
	for i := range templates {
		catQuery := `
			SELECT id, template_id, category, allocated_pct, limit_amount
			FROM template_categories
			WHERE template_id = $1
		`
		catRows, err := s.pool.Query(ctx, catQuery, templates[i].ID)
		if err != nil {
			return nil, err
		}
		for catRows.Next() {
			var cat models.TemplateCategory
			if err := catRows.Scan(&cat.ID, &cat.TemplateID, &cat.Category, &cat.AllocatedPct, &cat.LimitAmount); err != nil {
				catRows.Close()
				return nil, err
			}
			templates[i].Categories = append(templates[i].Categories, cat)
		}
		catRows.Close()
	}

	return templates, nil
}

func (s *Store) DeleteBudgetTemplate(ctx context.Context, userID string, templateID uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM budget_templates WHERE id = $1 AND user_id = $2`, templateID, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("template not found")
	}
	return nil
}

// Helper for pagination
func Paginate(totalCount, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(totalCount) / float64(pageSize)))
}
