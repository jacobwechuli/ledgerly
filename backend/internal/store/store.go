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

// --- Accounts ---

func (s *Store) CreateAccount(ctx context.Context, userID string, req models.CreateAccountRequest) (*models.Account, error) {
	account := &models.Account{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Type:        req.Type,
		SubType:     req.SubType,
		Currency:    req.Currency,
		Balance:     req.Balance,
		Institution: req.Institution,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO accounts (id, user_id, name, type, sub_type, currency, balance, institution, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, user_id, name, type, sub_type, currency, balance, institution, is_active, created_at, updated_at
	`

	err := s.pool.QueryRow(ctx, query,
		account.ID, account.UserID, account.Name, account.Type, account.SubType,
		account.Currency, account.Balance, account.Institution, account.IsActive,
		account.CreatedAt, account.UpdatedAt,
	).Scan(&account.ID, &account.UserID, &account.Name, &account.Type, &account.SubType,
		&account.Currency, &account.Balance, &account.Institution, &account.IsActive,
		&account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

func (s *Store) GetAccounts(ctx context.Context, userID string) ([]models.Account, error) {
	query := `
		SELECT id, user_id, name, type, sub_type, currency, balance, institution, is_active, created_at, updated_at
		FROM accounts
		WHERE user_id = $1 AND is_active = true
		ORDER BY type, name
	`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.SubType,
			&a.Currency, &a.Balance, &a.Institution, &a.IsActive,
			&a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}

	return accounts, nil
}

func (s *Store) UpdateAccountBalance(ctx context.Context, userID string, accountID uuid.UUID, newBalance float64) error {
	query := `UPDATE accounts SET balance = $3, updated_at = $4 WHERE id = $1 AND user_id = $2`
	result, err := s.pool.Exec(ctx, query, accountID, userID, newBalance, time.Now())
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("account not found")
	}
	return nil
}

func (s *Store) DeleteAccount(ctx context.Context, userID string, accountID uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `UPDATE accounts SET is_active = false, updated_at = $3 WHERE id = $1 AND user_id = $2`, accountID, userID, time.Now())
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("account not found")
	}
	return nil
}

// --- Transactions ---

func (s *Store) CreateTransaction(ctx context.Context, userID string, req models.CreateTransactionRequest) (*models.Transaction, error) {
	tx := &models.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		AccountID:   req.AccountID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Type:        req.Type,
		Category:    req.Category,
		Description: req.Description,
		Date:        req.Date,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if tx.Date.IsZero() {
		tx.Date = time.Now()
	}

	query := `
		INSERT INTO transactions (id, user_id, account_id, amount, currency, type, category, description, date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, user_id, account_id, amount, currency, type, category, description, date, created_at, updated_at
	`

	err := s.pool.QueryRow(ctx, query,
		tx.ID, tx.UserID, tx.AccountID, tx.Amount, tx.Currency, tx.Type,
		tx.Category, tx.Description, tx.Date, tx.CreatedAt, tx.UpdatedAt,
	).Scan(&tx.ID, &tx.UserID, &tx.AccountID, &tx.Amount, &tx.Currency,
		&tx.Type, &tx.Category, &tx.Description, &tx.Date, &tx.CreatedAt, &tx.UpdatedAt)

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
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE user_id = $1`, userID).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, account_id, amount, currency, type, category, description, date, created_at, updated_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.AccountID, &t.Amount, &t.Currency,
			&t.Type, &t.Category, &t.Description, &t.Date, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, t)
	}

	return transactions, totalCount, nil
}

func (s *Store) DeleteTransaction(ctx context.Context, userID string, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM transactions WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("transaction not found")
	}
	return nil
}

// --- Investments ---

func (s *Store) CreateInvestment(ctx context.Context, userID string, req models.CreateInvestmentRequest) (*models.Investment, error) {
	inv := &models.Investment{
		ID:             uuid.New(),
		UserID:         userID,
		AccountID:      req.AccountID,
		Symbol:         req.Symbol,
		Name:           req.Name,
		Type:           req.Type,
		Quantity:       req.Quantity,
		AvgCostBasis:   req.AvgCostBasis,
		CurrentPrice:   req.CurrentPrice,
		Currency:       req.Currency,
		UnrealizedGain: (req.CurrentPrice - req.AvgCostBasis) * req.Quantity,
		LastUpdated:    time.Now(),
		CreatedAt:      time.Now(),
	}

	query := `
		INSERT INTO investments (id, user_id, account_id, symbol, name, type, quantity, avg_cost_basis, current_price, currency, unrealized_gain, last_updated, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, user_id, account_id, symbol, name, type, quantity, avg_cost_basis, current_price, currency, unrealized_gain, last_updated, created_at
	`

	err := s.pool.QueryRow(ctx, query,
		inv.ID, inv.UserID, inv.AccountID, inv.Symbol, inv.Name, inv.Type,
		inv.Quantity, inv.AvgCostBasis, inv.CurrentPrice, inv.Currency,
		inv.UnrealizedGain, inv.LastUpdated, inv.CreatedAt,
	).Scan(&inv.ID, &inv.UserID, &inv.AccountID, &inv.Symbol, &inv.Name, &inv.Type,
		&inv.Quantity, &inv.AvgCostBasis, &inv.CurrentPrice, &inv.Currency,
		&inv.UnrealizedGain, &inv.LastUpdated, &inv.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create investment: %w", err)
	}

	return inv, nil
}

func (s *Store) GetInvestments(ctx context.Context, userID string) ([]models.Investment, error) {
	query := `
		SELECT id, user_id, account_id, symbol, name, type, quantity, avg_cost_basis, current_price, currency, unrealized_gain, last_updated, created_at
		FROM investments
		WHERE user_id = $1
		ORDER BY type, symbol
	`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var investments []models.Investment
	for rows.Next() {
		var i models.Investment
		if err := rows.Scan(&i.ID, &i.UserID, &i.AccountID, &i.Symbol, &i.Name, &i.Type,
			&i.Quantity, &i.AvgCostBasis, &i.CurrentPrice, &i.Currency,
			&i.UnrealizedGain, &i.LastUpdated, &i.CreatedAt); err != nil {
			return nil, err
		}
		investments = append(investments, i)
	}

	return investments, nil
}

func (s *Store) DeleteInvestment(ctx context.Context, userID string, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM investments WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("investment not found")
	}
	return nil
}

// --- Properties ---

func (s *Store) CreateProperty(ctx context.Context, userID string, req models.CreatePropertyRequest) (*models.Property, error) {
	prop := &models.Property{
		ID:            uuid.New(),
		UserID:        userID,
		AccountID:     req.AccountID,
		Name:          req.Name,
		Type:          req.Type,
		Description:   req.Description,
		CurrentValue:  req.CurrentValue,
		PurchasePrice: req.PurchasePrice,
		Currency:      req.Currency,
		Location:      req.Location,
		PurchaseDate:  req.PurchaseDate,
		LastValued:    time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	query := `
		INSERT INTO properties (id, user_id, account_id, name, type, description, current_value, purchase_price, currency, location, purchase_date, last_valued, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, user_id, account_id, name, type, description, current_value, purchase_price, currency, location, purchase_date, last_valued, created_at, updated_at
	`

	err := s.pool.QueryRow(ctx, query,
		prop.ID, prop.UserID, prop.AccountID, prop.Name, prop.Type, prop.Description,
		prop.CurrentValue, prop.PurchasePrice, prop.Currency, prop.Location,
		prop.PurchaseDate, prop.LastValued, prop.CreatedAt, prop.UpdatedAt,
	).Scan(&prop.ID, &prop.UserID, &prop.AccountID, &prop.Name, &prop.Type,
		&prop.Description, &prop.CurrentValue, &prop.PurchasePrice, &prop.Currency,
		&prop.Location, &prop.PurchaseDate, &prop.LastValued, &prop.CreatedAt, &prop.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create property: %w", err)
	}

	return prop, nil
}

func (s *Store) GetProperties(ctx context.Context, userID string) ([]models.Property, error) {
	query := `
		SELECT id, user_id, account_id, name, type, description, current_value, purchase_price, currency, location, purchase_date, last_valued, created_at, updated_at
		FROM properties
		WHERE user_id = $1
		ORDER BY type, name
	`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var properties []models.Property
	for rows.Next() {
		var p models.Property
		if err := rows.Scan(&p.ID, &p.UserID, &p.AccountID, &p.Name, &p.Type,
			&p.Description, &p.CurrentValue, &p.PurchasePrice, &p.Currency,
			&p.Location, &p.PurchaseDate, &p.LastValued, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		properties = append(properties, p)
	}

	return properties, nil
}

func (s *Store) DeleteProperty(ctx context.Context, userID string, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM properties WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("property not found")
	}
	return nil
}

// --- Net Worth ---

func (s *Store) GetNetWorthSummary(ctx context.Context, userID string) (*models.NetWorthSummary, error) {
	// Calculate cash (bank + mobile money accounts)
	var cashUSD float64
	cashQuery := `
		SELECT COALESCE(SUM(a.balance * c.rate_to_usd), 0)
		FROM accounts a
		JOIN currencies c ON a.currency = c.code
		WHERE a.user_id = $1 AND a.type IN ('bank', 'mobile_money') AND a.is_active = true
	`
	err := s.pool.QueryRow(ctx, cashQuery, userID).Scan(&cashUSD)
	if err != nil {
		return nil, err
	}

	// Calculate investments
	var investUSD float64
	investQuery := `
		SELECT COALESCE(SUM(i.quantity * i.current_price * c.rate_to_usd), 0)
		FROM investments i
		JOIN currencies c ON i.currency = c.code
		WHERE i.user_id = $1
	`
	err = s.pool.QueryRow(ctx, investQuery, userID).Scan(&investUSD)
	if err != nil {
		return nil, err
	}

	// Calculate property
	var propUSD float64
	propQuery := `
		SELECT COALESCE(SUM(p.current_value * c.rate_to_usd), 0)
		FROM properties p
		JOIN currencies c ON p.currency = c.code
		WHERE p.user_id = $1
	`
	err = s.pool.QueryRow(ctx, propQuery, userID).Scan(&propUSD)
	if err != nil {
		return nil, err
	}

	totalUSD := cashUSD + investUSD + propUSD

	// Calculate changes (compare to last snapshot)
	var prevTotal float64
	prevQuery := `
		SELECT total_usd FROM net_worth_snapshots
		WHERE user_id = $1 AND date < CURRENT_DATE
		ORDER BY date DESC LIMIT 1
	`
	_ = s.pool.QueryRow(ctx, prevQuery, userID).Scan(&prevTotal)

	change30d := totalUSD - prevTotal
	changePct30d := 0.0
	if prevTotal > 0 {
		changePct30d = (change30d / prevTotal) * 100
	}

	return &models.NetWorthSummary{
		TotalUSD:       totalUSD,
		CashUSD:        cashUSD,
		InvestmentsUSD: investUSD,
		PropertyUSD:    propUSD,
		Change30d:      change30d,
		ChangePct30d:   changePct30d,
	}, nil
}

func (s *Store) GetNetWorthHistory(ctx context.Context, userID string, days int) ([]models.NetWorthSnapshot, error) {
	if days <= 0 {
		days = 30
	}

	// Compute the cutoff date in Go to avoid any SQL type ambiguity.
	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	query := `
		SELECT id, user_id, date, total_usd, cash_usd, invest_usd, prop_usd, created_at
		FROM net_worth_snapshots
		WHERE user_id = $1 AND date >= $2
		ORDER BY date ASC
	`

	rows, err := s.pool.Query(ctx, query, userID, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snapshots := []models.NetWorthSnapshot{}
	for rows.Next() {
		var snap models.NetWorthSnapshot
		if err := rows.Scan(&snap.ID, &snap.UserID, &snap.Date, &snap.TotalUSD,
			&snap.CashUSD, &snap.InvestUSD, &snap.PropUSD, &snap.CreatedAt); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snap)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return snapshots, nil
}

func (s *Store) CreateNetWorthSnapshot(ctx context.Context, userID string) error {
	summary, err := s.GetNetWorthSummary(ctx, userID)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO net_worth_snapshots (id, user_id, date, total_usd, cash_usd, invest_usd, prop_usd, created_at)
		VALUES ($1, $2, CURRENT_DATE, $3, $4, $5, $6, NOW())
		ON CONFLICT (user_id, date) DO UPDATE SET
			total_usd = $3, cash_usd = $4, invest_usd = $5, prop_usd = $6
	`

	_, err = s.pool.Exec(ctx, query, uuid.New(), userID, summary.TotalUSD, summary.CashUSD, summary.InvestmentsUSD, summary.PropertyUSD)
	return err
}

// --- Currencies ---

func (s *Store) GetCurrency(ctx context.Context, code string) (*models.Currency, error) {
	query := `SELECT code, name, symbol, rate_to_usd, updated_at FROM currencies WHERE code = $1`
	var c models.Currency
	err := s.pool.QueryRow(ctx, query, code).Scan(&c.Code, &c.Name, &c.Symbol, &c.RateToUSD, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) GetAllCurrencies(ctx context.Context) ([]models.Currency, error) {
	query := `SELECT code, name, symbol, rate_to_usd, updated_at FROM currencies ORDER BY code`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currencies []models.Currency
	for rows.Next() {
		var c models.Currency
		if err := rows.Scan(&c.Code, &c.Name, &c.Symbol, &c.RateToUSD, &c.UpdatedAt); err != nil {
			return nil, err
		}
		currencies = append(currencies, c)
	}

	return currencies, nil
}

func (s *Store) ConvertCurrency(ctx context.Context, from, to string, amount float64) (*models.CurrencyConversion, error) {
	var fromRate, toRate float64

	err := s.pool.QueryRow(ctx, `SELECT rate_to_usd FROM currencies WHERE code = $1`, from).Scan(&fromRate)
	if err != nil {
		return nil, fmt.Errorf("currency %s not found", from)
	}

	err = s.pool.QueryRow(ctx, `SELECT rate_to_usd FROM currencies WHERE code = $1`, to).Scan(&toRate)
	if err != nil {
		return nil, fmt.Errorf("currency %s not found", to)
	}

	// Convert: amount in 'from' -> USD -> 'to'
	amountUSD := amount * fromRate
	result := amountUSD / toRate
	rate := fromRate / toRate

	return &models.CurrencyConversion{
		FromCurrency: from,
		ToCurrency:   to,
		Amount:       amount,
		Result:       result,
		Rate:         rate,
	}, nil
}

// --- Bills ---

func (s *Store) CreateBill(ctx context.Context, userID string, req models.CreateBillRequest) (*models.Bill, error) {
	bill := &models.Bill{
		ID:         uuid.New(),
		UserID:     userID,
		AccountID:  req.AccountID,
		Name:       req.Name,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Category:   req.Category,
		DueDate:    req.DueDate,
		Recurrence: req.Recurrence,
		IsPaid:     false,
		Notes:      req.Notes,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	query := `
		INSERT INTO bills (id, user_id, account_id, name, amount, currency, category, due_date, recurrence, is_paid, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, user_id, account_id, name, amount, currency, category, due_date, recurrence, is_paid, paid_date, notes, created_at, updated_at
	`

	err := s.pool.QueryRow(ctx, query,
		bill.ID, bill.UserID, bill.AccountID, bill.Name, bill.Amount, bill.Currency,
		bill.Category, bill.DueDate, bill.Recurrence, bill.IsPaid, bill.Notes,
		bill.CreatedAt, bill.UpdatedAt,
	).Scan(&bill.ID, &bill.UserID, &bill.AccountID, &bill.Name, &bill.Amount, &bill.Currency,
		&bill.Category, &bill.DueDate, &bill.Recurrence, &bill.IsPaid, &bill.PaidDate,
		&bill.Notes, &bill.CreatedAt, &bill.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create bill: %w", err)
	}

	return bill, nil
}

func (s *Store) GetBills(ctx context.Context, userID string, paid *bool) ([]models.Bill, error) {
	query := `
		SELECT id, user_id, account_id, name, amount, currency, category, due_date, recurrence, is_paid, paid_date, notes, created_at, updated_at
		FROM bills WHERE user_id = $1
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
		if err := rows.Scan(&b.ID, &b.UserID, &b.AccountID, &b.Name, &b.Amount, &b.Currency,
			&b.Category, &b.DueDate, &b.Recurrence, &b.IsPaid, &b.PaidDate,
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
		UPDATE bills SET is_paid = true, paid_date = $3, updated_at = $3
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, account_id, name, amount, currency, category, due_date, recurrence, is_paid, paid_date, notes, created_at, updated_at
	`

	var b models.Bill
	err := s.pool.QueryRow(ctx, query, billID, userID, now).Scan(
		&b.ID, &b.UserID, &b.AccountID, &b.Name, &b.Amount, &b.Currency,
		&b.Category, &b.DueDate, &b.Recurrence, &b.IsPaid, &b.PaidDate,
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

// --- Budget Goals ---

func (s *Store) CreateBudgetGoal(ctx context.Context, userID string, req models.CreateBudgetGoalRequest) (*models.BudgetGoal, error) {
	goal := &models.BudgetGoal{
		ID:            uuid.New(),
		UserID:        userID,
		Name:          req.Name,
		TargetAmount:  req.TargetAmount,
		CurrentAmount: 0,
		Currency:      req.Currency,
		Category:      req.Category,
		Deadline:      req.Deadline,
		Status:        "active",
		Priority:      req.Priority,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	query := `
		INSERT INTO budget_goals (id, user_id, name, target_amount, current_amount, currency, category, deadline, status, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, user_id, name, target_amount, current_amount, currency, category, deadline, status, priority, created_at, updated_at
	`

	err := s.pool.QueryRow(ctx, query,
		goal.ID, goal.UserID, goal.Name, goal.TargetAmount, goal.CurrentAmount,
		goal.Currency, goal.Category, goal.Deadline, goal.Status, goal.Priority,
		goal.CreatedAt, goal.UpdatedAt,
	).Scan(&goal.ID, &goal.UserID, &goal.Name, &goal.TargetAmount, &goal.CurrentAmount,
		&goal.Currency, &goal.Category, &goal.Deadline, &goal.Status, &goal.Priority,
		&goal.CreatedAt, &goal.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create budget goal: %w", err)
	}

	return goal, nil
}

func (s *Store) GetBudgetGoals(ctx context.Context, userID string) ([]models.BudgetGoal, error) {
	query := `
		SELECT id, user_id, name, target_amount, current_amount, currency, category, deadline, status, priority, created_at, updated_at
		FROM budget_goals WHERE user_id = $1 ORDER BY priority DESC, created_at DESC
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
			&g.Currency, &g.Category, &g.Deadline, &g.Status, &g.Priority,
			&g.CreatedAt, &g.UpdatedAt); err != nil {
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
		RETURNING id, user_id, name, target_amount, current_amount, currency, category, deadline, status, priority, created_at, updated_at
	`

	var g models.BudgetGoal
	err := s.pool.QueryRow(ctx, query, goalID, userID, amount, time.Now()).Scan(
		&g.ID, &g.UserID, &g.Name, &g.TargetAmount, &g.CurrentAmount,
		&g.Currency, &g.Category, &g.Deadline, &g.Status, &g.Priority,
		&g.CreatedAt, &g.UpdatedAt)

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

// --- Reports ---

func (s *Store) GetExpenditureReport(ctx context.Context, userID string, startDate, endDate time.Time, currency string) (*models.ExpenditureReport, error) {
	// If currency specified, convert all amounts to that currency
	var query string
	var args []interface{}

	if currency != "" {
		query = `
			SELECT t.category, SUM(t.amount * c.rate_to_usd / tc.rate_to_usd) as total, COUNT(*) as count
			FROM transactions t
			JOIN currencies c ON t.currency = c.code
			JOIN currencies tc ON tc.code = $4
			WHERE t.user_id = $1 AND t.type = 'expense' AND t.date >= $2 AND t.date <= $3
			GROUP BY t.category ORDER BY total DESC
		`
		args = []interface{}{userID, startDate, endDate, currency}
	} else {
		query = `
			SELECT category, SUM(amount) as total, COUNT(*) as count
			FROM transactions
			WHERE user_id = $1 AND type = 'expense' AND date >= $2 AND date <= $3
			GROUP BY category ORDER BY total DESC
		`
		args = []interface{}{userID, startDate, endDate}
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := &models.ExpenditureReport{
		StartDate: startDate,
		EndDate:   endDate,
		Currency:  currency,
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
		SELECT type, COALESCE(SUM(amount * c.rate_to_usd), 0)
		FROM transactions t
		JOIN currencies c ON t.currency = c.code
		WHERE t.user_id = $1
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
		FROM budget_templates WHERE user_id = $1 OR is_public = true ORDER BY created_at DESC
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

	for i := range templates {
		catRows, err := s.pool.Query(ctx, `SELECT id, template_id, category, allocated_pct, limit_amount FROM template_categories WHERE template_id = $1`, templates[i].ID)
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

// --- User Settings ---

func (s *Store) GetUserSettings(ctx context.Context, userID string) (*models.UserSettings, error) {
	query := `SELECT user_id, privacy_mode, default_currency, theme, updated_at FROM user_settings WHERE user_id = $1`
	var settings models.UserSettings
	err := s.pool.QueryRow(ctx, query, userID).Scan(&settings.UserID, &settings.PrivacyMode, &settings.DefaultCurrency, &settings.Theme, &settings.UpdatedAt)
	if err != nil {
		// Return defaults if not found
		return &models.UserSettings{
			UserID:          userID,
			PrivacyMode:     false,
			DefaultCurrency: "USD",
			Theme:           "system",
			UpdatedAt:       time.Now(),
		}, nil
	}
	return &settings, nil
}

func (s *Store) UpdateUserSettings(ctx context.Context, userID string, req models.UpdateUserSettingsRequest) (*models.UserSettings, error) {
	current, err := s.GetUserSettings(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.PrivacyMode != nil {
		current.PrivacyMode = *req.PrivacyMode
	}
	if req.DefaultCurrency != nil {
		current.DefaultCurrency = *req.DefaultCurrency
	}
	if req.Theme != nil {
		current.Theme = *req.Theme
	}
	current.UpdatedAt = time.Now()

	query := `
		INSERT INTO user_settings (user_id, privacy_mode, default_currency, theme, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			privacy_mode = $2, default_currency = $3, theme = $4, updated_at = $5
	`

	_, err = s.pool.Exec(ctx, query, userID, current.PrivacyMode, current.DefaultCurrency, current.Theme, current.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return current, nil
}

// --- Tax Estimates ---

func (s *Store) EstimateCapitalGains(ctx context.Context, userID string, year int) (float64, error) {
	// Sum unrealized gains from investments
	query := `
		SELECT COALESCE(SUM(unrealized_gain * c.rate_to_usd), 0)
		FROM investments i
		JOIN currencies c ON i.currency = c.code
		WHERE i.user_id = $1
	`

	var totalGains float64
	err := s.pool.QueryRow(ctx, query, userID).Scan(&totalGains)
	if err != nil {
		return 0, err
	}

	return totalGains, nil
}

func (s *Store) GetTaxRecords(ctx context.Context, userID string, year int) ([]models.TaxRecord, error) {
	query := `
		SELECT id, user_id, year, type, amount, currency, description, estimated_at, created_at
		FROM tax_records WHERE user_id = $1 AND year = $2 ORDER BY type
	`

	rows, err := s.pool.Query(ctx, query, userID, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.TaxRecord
	for rows.Next() {
		var r models.TaxRecord
		if err := rows.Scan(&r.ID, &r.UserID, &r.Year, &r.Type, &r.Amount, &r.Currency, &r.Description, &r.EstimatedAt, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	return records, nil
}

// --- Currency Exposure Analysis ---

func (s *Store) GetCurrencyExposure(ctx context.Context, userID string) (map[string]float64, error) {
	query := `
		SELECT t.currency, SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE -t.amount END) * c.rate_to_usd
		FROM transactions t
		JOIN currencies c ON t.currency = c.code
		WHERE t.user_id = $1 AND t.date >= CURRENT_DATE - INTERVAL '30 days'
		GROUP BY t.currency
	`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exposure := make(map[string]float64)
	for rows.Next() {
		var currency string
		var amount float64
		if err := rows.Scan(&currency, &amount); err != nil {
			return nil, err
		}
		exposure[currency] = amount
	}

	return exposure, nil
}

// Helper for pagination
func Paginate(totalCount, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(totalCount) / float64(pageSize)))
}
