-- Ledgerly Database Schema
-- Migration: 001_initial_schema
-- Targets high-net-worth users with multi-currency, multi-account, investment, and property tracking

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Currencies table (exchange rates)
CREATE TABLE IF NOT EXISTS currencies (
    code VARCHAR(3) PRIMARY KEY, -- ISO 4217
    name VARCHAR(100) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    rate_to_usd DECIMAL(20,8) NOT NULL, -- 1 unit of this currency = X USD
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Seed common currencies
INSERT INTO currencies (code, name, symbol, rate_to_usd) VALUES
    ('USD', 'US Dollar', '$', 1.0),
    ('KES', 'Kenyan Shilling', 'KSh', 0.0077),
    ('GBP', 'British Pound', '£', 1.27),
    ('EUR', 'Euro', '€', 1.08),
    ('CHF', 'Swiss Franc', 'CHF', 1.13),
    ('JPY', 'Japanese Yen', '¥', 0.0067),
    ('ZAR', 'South African Rand', 'R', 0.054),
    ('NGN', 'Nigerian Naira', '₦', 0.00065),
    ('AED', 'UAE Dirham', 'د.إ', 0.27),
    ('SGD', 'Singapore Dollar', 'S$', 0.74)
ON CONFLICT (code) DO NOTHING;

-- User settings table
CREATE TABLE IF NOT EXISTS user_settings (
    user_id VARCHAR(255) PRIMARY KEY,
    privacy_mode BOOLEAN NOT NULL DEFAULT false,
    default_currency VARCHAR(3) NOT NULL DEFAULT 'USD' REFERENCES currencies(code),
    theme VARCHAR(20) NOT NULL DEFAULT 'system' CHECK (theme IN ('light', 'dark', 'system')),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Accounts table (multi-account support: bank, mobile money, investment, property)
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('bank', 'mobile_money', 'investment', 'property')),
    sub_type VARCHAR(50) NOT NULL, -- checking, savings, mpesa, stock_portfolio, crypto_wallet, real_estate, vehicle
    currency VARCHAR(3) NOT NULL REFERENCES currencies(code),
    balance DECIMAL(20,4) NOT NULL DEFAULT 0,
    institution VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_accounts_type ON accounts(type);
CREATE INDEX idx_accounts_currency ON accounts(currency);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    amount DECIMAL(20,4) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL REFERENCES currencies(code),
    type VARCHAR(20) NOT NULL CHECK (type IN ('income', 'expense', 'transfer')),
    category VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_account_id ON transactions(account_id);
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_category ON transactions(category);
CREATE INDEX idx_transactions_user_date ON transactions(user_id, date DESC);
CREATE INDEX idx_transactions_currency ON transactions(currency);

-- Investments table (stocks, funds, crypto)
CREATE TABLE IF NOT EXISTS investments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(30) NOT NULL CHECK (type IN ('stock', 'etf', 'mutual_fund', 'crypto', 'bond')),
    quantity DECIMAL(20,8) NOT NULL CHECK (quantity >= 0),
    avg_cost_basis DECIMAL(20,8) NOT NULL DEFAULT 0,
    current_price DECIMAL(20,8) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL REFERENCES currencies(code),
    unrealized_gain DECIMAL(20,4) NOT NULL DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_investments_user_id ON investments(user_id);
CREATE INDEX idx_investments_account_id ON investments(account_id);
CREATE INDEX idx_investments_type ON investments(type);
CREATE INDEX idx_investments_symbol ON investments(symbol);

-- Properties table (real estate, vehicles, art, etc.)
CREATE TABLE IF NOT EXISTS properties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(30) NOT NULL CHECK (type IN ('real_estate', 'vehicle', 'art', 'jewelry', 'other')),
    description TEXT DEFAULT '',
    current_value DECIMAL(20,4) NOT NULL DEFAULT 0,
    purchase_price DECIMAL(20,4) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL REFERENCES currencies(code),
    location VARCHAR(500),
    purchase_date TIMESTAMP WITH TIME ZONE,
    last_valued TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_properties_user_id ON properties(user_id);
CREATE INDEX idx_properties_account_id ON properties(account_id);
CREATE INDEX idx_properties_type ON properties(type);

-- Net worth snapshots (tracked over time)
CREATE TABLE IF NOT EXISTS net_worth_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    total_usd DECIMAL(20,4) NOT NULL DEFAULT 0,
    cash_usd DECIMAL(20,4) NOT NULL DEFAULT 0,
    invest_usd DECIMAL(20,4) NOT NULL DEFAULT 0,
    prop_usd DECIMAL(20,4) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, date)
);

CREATE INDEX idx_net_worth_snapshots_user_date ON net_worth_snapshots(user_id, date DESC);

-- Bills table
CREATE TABLE IF NOT EXISTS bills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id),
    name VARCHAR(255) NOT NULL,
    amount DECIMAL(20,4) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL REFERENCES currencies(code),
    category VARCHAR(100) NOT NULL,
    due_date TIMESTAMP WITH TIME ZONE NOT NULL,
    recurrence VARCHAR(20) NOT NULL CHECK (recurrence IN ('monthly', 'weekly', 'yearly', 'once')),
    is_paid BOOLEAN NOT NULL DEFAULT false,
    paid_date TIMESTAMP WITH TIME ZONE,
    notes TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bills_user_id ON bills(user_id);
CREATE INDEX idx_bills_due_date ON bills(due_date);
CREATE INDEX idx_bills_is_paid ON bills(is_paid);
CREATE INDEX idx_bills_user_due ON bills(user_id, due_date);

-- Budget goals (large goals: property, education, retirement, etc.)
CREATE TABLE IF NOT EXISTS budget_goals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    target_amount DECIMAL(20,4) NOT NULL CHECK (target_amount > 0),
    current_amount DECIMAL(20,4) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL REFERENCES currencies(code),
    category VARCHAR(50) NOT NULL CHECK (category IN ('property', 'education', 'retirement', 'travel', 'business')),
    deadline TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'paused')),
    priority VARCHAR(10) NOT NULL DEFAULT 'medium' CHECK (priority IN ('high', 'medium', 'low')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_budget_goals_user_id ON budget_goals(user_id);
CREATE INDEX idx_budget_goals_status ON budget_goals(status);
CREATE INDEX idx_budget_goals_priority ON budget_goals(priority);

-- Budget templates
CREATE TABLE IF NOT EXISTS budget_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    is_public BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_budget_templates_user_id ON budget_templates(user_id);

-- Template categories
CREATE TABLE IF NOT EXISTS template_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id UUID NOT NULL REFERENCES budget_templates(id) ON DELETE CASCADE,
    category VARCHAR(100) NOT NULL,
    allocated_pct DECIMAL(5,2) NOT NULL CHECK (allocated_pct >= 0 AND allocated_pct <= 100),
    limit_amount DECIMAL(20,4)
);

CREATE INDEX idx_template_categories_template_id ON template_categories(template_id);

-- Tax records (capital gains, income tax estimates)
CREATE TABLE IF NOT EXISTS tax_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    year INTEGER NOT NULL,
    type VARCHAR(30) NOT NULL CHECK (type IN ('capital_gains', 'income_tax', 'dividend_tax')),
    amount DECIMAL(20,4) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL REFERENCES currencies(code),
    description TEXT DEFAULT '',
    estimated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tax_records_user_year ON tax_records(user_id, year);
CREATE INDEX idx_tax_records_type ON tax_records(type);

-- Currency exchange rate history (for audit)
CREATE TABLE IF NOT EXISTS exchange_rate_history (
    id BIGSERIAL PRIMARY KEY,
    from_currency VARCHAR(3) NOT NULL REFERENCES currencies(code),
    to_currency VARCHAR(3) NOT NULL REFERENCES currencies(code),
    rate DECIMAL(20,8) NOT NULL,
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exchange_rate_history_currencies ON exchange_rate_history(from_currency, to_currency, recorded_at DESC);
