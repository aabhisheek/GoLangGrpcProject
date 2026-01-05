-- Voucher Payment Service Database Schema
-- MySQL 8.0+

-- Create database if not exists
CREATE DATABASE IF NOT EXISTS voucher_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE voucher_db;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Wallets table
CREATE TABLE IF NOT EXISTS wallets (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL UNIQUE,
    balance DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'INR',
    is_locked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    CHECK (balance >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Vouchers table (available vouchers in the system)
CREATE TABLE IF NOT EXISTS vouchers (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    brand VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,
    face_value DECIMAL(10, 2) NOT NULL,
    discount_percentage DECIMAL(5, 2) DEFAULT 0.00,
    selling_price DECIMAL(10, 2) NOT NULL,
    stock_quantity INT NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    valid_from TIMESTAMP NOT NULL,
    valid_until TIMESTAMP NOT NULL,
    terms_and_conditions TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_brand (brand),
    INDEX idx_active (is_active),
    INDEX idx_price (selling_price),
    INDEX idx_valid_dates (valid_from, valid_until),
    FULLTEXT INDEX idx_search (name, description, brand)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Purchased vouchers table (user-owned vouchers)
CREATE TABLE IF NOT EXISTS purchased_vouchers (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    voucher_id VARCHAR(36) NOT NULL,
    transaction_id VARCHAR(36) NOT NULL,
    voucher_code VARCHAR(50) NOT NULL UNIQUE,
    pin VARCHAR(20) NOT NULL,
    status ENUM('active', 'used', 'expired', 'refunded') DEFAULT 'active',
    purchase_price DECIMAL(10, 2) NOT NULL,
    valid_until TIMESTAMP NOT NULL,
    used_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (voucher_id) REFERENCES vouchers(id) ON DELETE RESTRICT,
    INDEX idx_user_id (user_id),
    INDEX idx_voucher_id (voucher_id),
    INDEX idx_transaction_id (transaction_id),
    INDEX idx_status (status),
    INDEX idx_voucher_code (voucher_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    type ENUM('credit', 'debit') NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    balance_before DECIMAL(15, 2) NOT NULL,
    balance_after DECIMAL(15, 2) NOT NULL,
    description VARCHAR(255),
    reference_id VARCHAR(100),
    payment_method VARCHAR(50),
    status ENUM('pending', 'completed', 'failed', 'refunded') DEFAULT 'pending',
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id),
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_reference_id (reference_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Payment events table (for idempotency and audit)
CREATE TABLE IF NOT EXISTS payment_events (
    id VARCHAR(36) PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    transaction_id VARCHAR(36),
    payload JSON NOT NULL,
    status ENUM('pending', 'processed', 'failed') DEFAULT 'pending',
    retry_count INT DEFAULT 0,
    processed_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_event_type (event_type),
    INDEX idx_status (status),
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert sample data

-- Insert sample users
INSERT INTO users (id, email, full_name, phone_number, status) VALUES
('user-001', 'john.doe@example.com', 'John Doe', '+919876543210', 'active'),
('user-002', 'jane.smith@example.com', 'Jane Smith', '+919876543211', 'active'),
('user-003', 'bob.wilson@example.com', 'Bob Wilson', '+919876543212', 'active');

-- Insert wallets for users
INSERT INTO wallets (id, user_id, balance, currency) VALUES
('wallet-001', 'user-001', 5000.00, 'INR'),
('wallet-002', 'user-002', 10000.00, 'INR'),
('wallet-003', 'user-003', 2500.00, 'INR');

-- Insert sample vouchers
INSERT INTO vouchers (id, name, description, brand, category, face_value, discount_percentage, selling_price, stock_quantity, valid_from, valid_until, terms_and_conditions) VALUES
('voucher-001', 'Amazon Gift Card ₹500', 'Amazon shopping voucher worth ₹500', 'Amazon', 'E-Commerce', 500.00, 5.00, 475.00, 100, NOW(), DATE_ADD(NOW(), INTERVAL 365 DAY), 'Valid for 1 year. Non-refundable.'),
('voucher-002', 'Flipkart Gift Card ₹1000', 'Flipkart shopping voucher worth ₹1000', 'Flipkart', 'E-Commerce', 1000.00, 7.50, 925.00, 50, NOW(), DATE_ADD(NOW(), INTERVAL 365 DAY), 'Valid for 1 year. Non-refundable.'),
('voucher-003', 'Starbucks Coffee Voucher ₹300', 'Enjoy your favorite coffee at Starbucks', 'Starbucks', 'Food & Beverage', 300.00, 10.00, 270.00, 200, NOW(), DATE_ADD(NOW(), INTERVAL 180 DAY), 'Valid at all Starbucks outlets. Valid for 6 months.'),
('voucher-004', 'BookMyShow Movie Voucher ₹500', 'Watch movies at any theater', 'BookMyShow', 'Entertainment', 500.00, 8.00, 460.00, 150, NOW(), DATE_ADD(NOW(), INTERVAL 90 DAY), 'Valid for 3 months. Cannot be clubbed with other offers.'),
('voucher-005', 'Zomato Food Voucher ₹250', 'Order food from your favorite restaurants', 'Zomato', 'Food & Beverage', 250.00, 12.00, 220.00, 300, NOW(), DATE_ADD(NOW(), INTERVAL 60 DAY), 'Valid for 60 days. Minimum order ₹300.'),
('voucher-006', 'Uber Rides Voucher ₹400', 'Get ₹400 off on Uber rides', 'Uber', 'Travel', 400.00, 15.00, 340.00, 75, NOW(), DATE_ADD(NOW(), INTERVAL 30 DAY), 'Valid for 30 days. Multiple rides allowed.'),
('voucher-007', 'Google Play Gift Card ₹500', 'Buy apps, games, and more', 'Google Play', 'Digital', 500.00, 5.00, 475.00, 500, NOW(), DATE_ADD(NOW(), INTERVAL 365 DAY), 'Valid for 1 year. Non-transferable.'),
('voucher-008', 'Netflix Subscription Voucher ₹199', '1 month Mobile plan subscription', 'Netflix', 'Entertainment', 199.00, 0.00, 199.00, 1000, NOW(), DATE_ADD(NOW(), INTERVAL 365 DAY), 'Activate within 1 year. Valid for 1 month after activation.'),
('voucher-009', 'Swiggy Food Voucher ₹500', 'Order delicious food on Swiggy', 'Swiggy', 'Food & Beverage', 500.00, 10.00, 450.00, 250, NOW(), DATE_ADD(NOW(), INTERVAL 90 DAY), 'Valid for 90 days. No minimum order.'),
('voucher-010', 'Big Bazaar Shopping Voucher ₹1000', 'Shop for groceries and more', 'Big Bazaar', 'Retail', 1000.00, 5.00, 950.00, 80, NOW(), DATE_ADD(NOW(), INTERVAL 180 DAY), 'Valid at all Big Bazaar stores. Valid for 6 months.');

-- Create indexes for better query performance
CREATE INDEX idx_vouchers_category_brand ON vouchers(category, brand);
CREATE INDEX idx_transactions_user_created ON transactions(user_id, created_at DESC);
CREATE INDEX idx_purchased_user_status ON purchased_vouchers(user_id, status);

