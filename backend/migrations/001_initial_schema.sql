-- Ledgerly Database Schema
-- Migration: 001_initial_schema

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    type VARCHAR(20) NOT NULL CHECK (type IN ('income', 'expense')),
    category VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    source VARCHAR(50) NOT NULL CHECK (source IN ('bank', 'cash', 'mobile_money', 'card')),
    date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_category ON transactions(category);
CREATE INDEX idx_transactions_user_date ON transactions(user_id, date DESC);

-- Mobile Money Transactions table
CREATE TABLE IF NOT EXISTS mobile_money_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    transaction_id VARCHAR(255) NOT NULL,
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    type VARCHAR(20) NOT NULL CHECK (type IN ('send', 'receive', 'paybill', 'buygoods')),
    phone_number VARCHAR(20) NOT NULL,
    provider VARCHAR(50) NOT NULL CHECK (provider IN ('mpesa', 'airtel_money', 'tigo_pesa')),
    counterparty_name VARCHAR(255) DEFAULT '',
    description TEXT DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'completed' CHECK (status IN ('completed', 'pending', 'failed')),
    date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mobile_money_user_id ON mobile_money_transactions(user_id);
CREATE INDEX idx_mobile_money_date ON mobile_money_transactions(date);
CREATE INDEX idx_mobile_money_provider ON mobile_money_transactions(provider);
CREATE INDEX idx_mobile_money_transaction_id ON mobile_money_transactions(transaction_id);

-- Budget Goals table
CREATE TABLE IF NOT EXISTS budget_goals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    target_amount DECIMAL(15,2) NOT NULL CHECK (target_amount > 0),
    current_amount DECIMAL(15,2) NOT NULL DEFAULT 0 CHECK (current_amount >= 0),
    category VARCHAR(100) NOT NULL,
    deadline TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'paused')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_budget_goals_user_id ON budget_goals(user_id);
CREATE INDEX idx_budget_goals_status ON budget_goals(status);

-- Bills table
CREATE TABLE IF NOT EXISTS bills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
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

-- Budget Templates table
CREATE TABLE IF NOT EXISTS budget_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    is_public BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_budget_templates_user_id ON budget_templates(user_id);
CREATE INDEX idx_budget_templates_public ON budget_templates(is_public) WHERE is_public = true;

-- Template Categories table
CREATE TABLE IF NOT EXISTS template_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    template_id UUID NOT NULL REFERENCES budget_templates(id) ON DELETE CASCADE,
    category VARCHAR(100) NOT NULL,
    allocated_pct DECIMAL(5,2) NOT NULL CHECK (allocated_pct >= 0 AND allocated_pct <= 100),
    limit_amount DECIMAL(15,2)
);

CREATE INDEX idx_template_categories_template_id ON template_categories(template_id);

-- Rate limiting table (for Redis fallback / audit)
CREATE TABLE IF NOT EXISTS rate_limit_audit (
    id BIGSERIAL PRIMARY KEY,
    identifier VARCHAR(255) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    request_count INTEGER NOT NULL DEFAULT 1,
    window_start TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rate_limit_audit_identifier ON rate_limit_audit(identifier, window_start);
