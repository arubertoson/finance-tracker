-- Initial Schema Migration
BEGIN TRANSACTION;

-- Core categories for transaction classification
CREATE TABLE categories (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    type TEXT CHECK(type IN ('income', 'expense')),
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default categories first (before creating transactions table)
-- These categories are predefined to help users classify their transactions
INSERT INTO categories (name, type, description) VALUES
    ('Uncategorized', 'expense', 'Transactions not yet categorized');

-- Transactions table with essential tracking
CREATE TABLE transactions (
    id INTEGER PRIMARY KEY,
    date DATE NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    description TEXT,
    category_id INTEGER NOT NULL DEFAULT 1,
    entity TEXT,
    recurring_pattern_id INTEGER,  -- If this is set, it's recurring
    FOREIGN KEY(category_id) REFERENCES categories(id),
    FOREIGN KEY(recurring_pattern_id) REFERENCES recurring_patterns(id)
);

CREATE TABLE recurring_patterns (
    id INTEGER PRIMARY KEY,
    frequency TEXT CHECK(frequency IN ('daily', 'weekly', 'monthly', 'yearly')),
    interval INTEGER NOT NULL DEFAULT 1,  -- how often to repeat (e.g., every 2 weeks)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Monthly budget targets
CREATE TABLE budget_plans (
    id INTEGER PRIMARY KEY,
    category_id INTEGER NOT NULL,
    month DATE NOT NULL,
    planned_amount DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(category_id) REFERENCES categories(id),
    UNIQUE(category_id, month),
    rollover BOOLEAN DEFAULT false
);

-- Savings goals
CREATE TABLE savings_goals (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    target_amount DECIMAL(10,2) NOT NULL,
    current_amount DECIMAL(10,2) DEFAULT 0,
    target_date DATE,
    is_completed BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Update category_rules table to be more flexible
DROP TABLE IF EXISTS category_rules;
CREATE TABLE category_rules (
    id INTEGER PRIMARY KEY,
    pattern TEXT NOT NULL,           -- Regex pattern or simple wildcard
    category_id INTEGER NOT NULL,
    priority INTEGER DEFAULT 0,      -- Higher priority rules are checked first
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(category_id) REFERENCES categories(id)
);

COMMIT;