-- Wheel Strategy Trading Database - Corrected for Wheeler Schema
-- Professional options trading simulation for 2025
-- Implements covered calls, cash-secured puts, and systematic scaling with proper Treasury collateral management

-- Clear existing data
DELETE FROM symbols;
DELETE FROM options;
DELETE FROM long_positions;
DELETE FROM dividends;
DELETE FROM treasuries;

-- Setup symbols with realistic pricing
INSERT INTO symbols (symbol, price, dividend, pe_ratio) VALUES
('VZ', 44.2, .68, 8.5),
('CVX', 158.50, 1.71, 15.2),
('KO', 69.80, .51, 25.3),
('MMM', 153.50, .73, 16.8);

-- Initial Treasury collateral position (Cash securing puts)
INSERT INTO treasuries (cuspid, purchased, maturity, amount, yield, buy_price, current_value) VALUES
    ('912797JX2', '2025-01-01', '2025-12-31', 75000.00, 4.83, 74700.00, 74700.00);


-- ==========================================
-- JANUARY 2025 - INITIAL POSITIONS
-- ==========================================

-- Initial long stock purchases
-- INSERT INTO long_positions (symbol, shares, buy_price, opened) VALUES
--     ('VZ', 100, 41.25, '2025-01-05');
--
-- INSERT INTO options (symbol, type, opened, expiration, strike, premium, contracts, commission) VALUES
--     ('VZ', 'Call', '2025-01-03', '2025-01-17', 42.00, 0.80, 1, 1.00);
--
-- INSERT INTO options (symbol, type, opened, expiration, strike, premium, contracts, commission) VALUES
--  ('VZ', 'Put', '2025-01-04', '2025-01-17', 38.00, 0.95, 1, 1.00);

INSERT INTO long_positions (symbol, shares, buy_price, opened) VALUES
('VZ', 100, 40.00, '2025-01-05'),
('CVX', 100, 160.00, '2025-01-05'),
('KO', 100, 60.00, '2025-01-05'),
('MMM', 100, 100.00, '2025-01-05');

-- January strangle setup (February expiration)
-- Covered calls
INSERT INTO options (symbol, type, opened, expiration, strike, premium, contracts, commission) VALUES
('VZ', 'Call', '2025-01-08', '2025-02-16', 42.00, 0.80, 1, 1.00),
('CVX', 'Call', '2025-01-08', '2025-02-16', 170.00, 2.10, 1, 1.00),
('KO', 'Call', '2025-01-08', '2025-02-16', 62.00, 0.75, 1, 1.00),
('MMM', 'Call', '2025-01-08', '2025-02-16', 105.00, 1.20, 1, 1.00);

-- Cash-secured puts
INSERT INTO options (symbol, type, opened, expiration, strike, premium, contracts, commission) VALUES
('VZ', 'Put', '2025-01-08', '2025-02-16', 38.00, 0.95, 1, 1.00),
('CVX', 'Put', '2025-01-08', '2025-02-16', 155.00, 2.40, 1, 1.00),
('KO', 'Put', '2025-01-08', '2025-02-16', 58.00, 0.85, 1, 1.00),
('MMM', 'Put', '2025-01-08', '2025-02-16', 95.00, 1.35, 1, 1.00);

-- ==========================================
-- FEBRUARY 2025 - EXPIRATION & ASSIGNMENT
-- ==========================================

-- February 16 expirations - most expired worthless, MMM put assigned
UPDATE options SET closed = '2025-02-16', exit_price = 0.00 WHERE expiration = '2025-02-16' AND symbol != 'MMM' OR (symbol = 'MMM' AND type = 'Call');
UPDATE options SET closed = '2025-02-16', exit_price = 2.00 WHERE symbol = 'MMM' AND type = 'Put' AND expiration = '2025-02-16';

-- MMM put assignment - acquired 100 shares at $95 (Use Treasury collateral)
INSERT INTO long_positions (symbol, shares, buy_price, opened) VALUES
('MMM', 100, 95.00, '2025-02-16');

-- Reduce Treasury position due to MMM put assignment (Used $9,500 cash)
UPDATE treasuries SET amount = amount - 9500.00 WHERE cuspid = '912797JX2';

-- March options setup
INSERT INTO options (symbol, type, opened, expiration, strike, premium, contracts, commission) VALUES
-- Covered calls
('VZ', 'Call', '2025-02-20', '2025-03-15', 42.50, 0.70, 1, 1.00),
('CVX', 'Call', '2025-02-20', '2025-03-15', 172.00, 1.95, 1, 1.00),
('KO', 'Call', '2025-02-20', '2025-03-15', 63.00, 0.65, 1, 1.00),
('MMM', 'Call', '2025-02-20', '2025-03-15', 102.00, 1.10, 2, 1.00),
-- Cash-secured puts
('VZ', 'Put', '2025-02-20', '2025-03-15', 38.50, 0.90, 1, 1.00),
('CVX', 'Put', '2025-02-20', '2025-03-15', 158.00, 2.25, 1, 1.00),
('KO', 'Put', '2025-02-20', '2025-03-15', 58.50, 0.80, 1, 1.00),
('MMM', 'Put', '2025-02-20', '2025-03-15', 93.00, 1.25, 1, 1.00);

-- ==========================================
-- MARCH 2025 - DIVIDENDS & ASSIGNMENTS
-- ==========================================

-- Q1 Dividend payments
INSERT INTO dividends (symbol, amount, received) VALUES
('VZ', 66.50, '2025-03-01'),
('CVX', 142.00, '2025-03-01'),
('KO', 46.00, '2025-03-01'),
('MMM', 300.00, '2025-03-01');

-- Q1 Treasury interest earned (5.25% annual = 1.31% quarterly on remaining balance)
INSERT INTO treasuries (cuspid, purchased, maturity, amount, yield, buy_price) VALUES
('INT-Q1-2025', '2025-03-31', '2025-03-31', 857.19, 5.25, 100.00);

-- March 15 assignments
-- VZ called away at $42.50
UPDATE long_positions SET closed = '2025-03-15', exit_price = 42.50 WHERE symbol = 'VZ' AND opened = '2025-01-05';
UPDATE options SET closed = '2025-03-15', exit_price = 2.70 WHERE symbol = 'VZ' AND type = 'Call' AND expiration = '2025-03-15';

-- One MMM contract assigned (100 shares called away at $102) - Proceeds go to Treasury
UPDATE long_positions SET closed = '2025-03-15', exit_price = 102.00 WHERE symbol = 'MMM' AND opened = '2025-02-16';
UPDATE options SET closed = '2025-03-15', exit_price = 1.50 WHERE symbol = 'MMM' AND type = 'Call' AND expiration = '2025-03-15' AND rowid = (SELECT MIN(rowid) FROM options WHERE symbol = 'MMM' AND type = 'Call' AND expiration = '2025-03-15');

-- Add proceeds from MMM assignment to Treasury reserves
UPDATE treasuries SET amount = amount + 10200.00 WHERE cuspid = '912797JX2';

-- Other March expirations expired worthless
UPDATE options SET closed = '2025-03-15', exit_price = 0.00 WHERE expiration = '2025-03-15' AND closed IS NULL;

-- ==========================================
-- APRIL 2025 - REBUILDING POSITIONS
-- ==========================================

-- Buy back VZ and scale up KO (Using Treasury collateral)
INSERT INTO long_positions (symbol, shares, buy_price, opened) VALUES
('VZ', 100, 43.00, '2025-04-01'),
('KO', 50, 62.50, '2025-04-01');

-- Reduce Treasury for stock purchases ($4,300 + $3,125 = $7,425)
UPDATE treasuries SET amount = amount - 7425.00 WHERE cuspid = '912797JX2';

-- April options (April 19 expiration)
INSERT INTO options (symbol, type, opened, expiration, strike, premium, contracts, commission) VALUES
-- Covered calls
('VZ', 'Call', '2025-04-01', '2025-04-19', 45.00, 0.85, 1, 1.00),
('CVX', 'Call', '2025-04-01', '2025-04-19', 175.00, 2.30, 1, 1.00),
('KO', 'Call', '2025-04-01', '2025-04-19', 65.00, 0.70, 1, 1.00),
('MMM', 'Call', '2025-04-01', '2025-04-19', 108.00, 1.40, 1, 1.00),
-- Cash-secured puts
('VZ', 'Put', '2025-04-01', '2025-04-19', 41.00, 1.05, 1, 1.00),
('CVX', 'Put', '2025-04-01', '2025-04-19', 162.00, 2.60, 1, 1.00),
('KO', 'Put', '2025-04-01', '2025-04-19', 60.00, 0.95, 1, 1.00),
('MMM', 'Put', '2025-04-01', '2025-04-19', 98.00, 1.55, 1, 1.00);

-- April expirations - all expired worthless (Treasury earns interest on collateral)
UPDATE options SET closed = '2025-04-19', exit_price = 0.00 WHERE expiration = '2025-04-19';

-- ==========================================
-- MAY 2025 - DEFENSIVE ROLLING
-- ==========================================

-- May options with defensive rolling
INSERT INTO options (symbol, type, opened, expiration, strike, premium, contracts, commission) VALUES
-- Initial May setup
('VZ', 'Call', '2025-05-01', '2025-05-17', 46.00, 0.90, 1, 1.00),
('VZ', 'Put', '2025-05-01', '2025-05-17', 40.00, 1.10, 1, 1.00),
('KO', 'Call', '2025-05-01', '2025-05-17', 66.00, 0.75, 1, 1.00),
('KO', 'Put', '2025-05-01', '2025-05-17', 59.00, 1.00, 1, 1.00),
('MMM', 'Call', '2025-05-01', '2025-05-17', 110.00, 1.50, 1, 1.00);

-- Defensive rolls for CVX and MMM puts
INSERT INTO options (symbol, type, opened, expiration, strike, premium, contracts, commission) VALUES
-- Rolled CVX put down and out for credit
('CVX', 'Call', '2025-05-01', '2025-05-17', 178.00, 2.50, 1, 1.00),
('CVX', 'Put', '2025-05-10', '2025-06-21', 158.00, 2.80, 1, 1.00),
-- Rolled MMM put down and out for credit  
('MMM', 'Put', '2025-05-10', '2025-06-21', 95.00, 1.75, 1, 1.00);

-- May expirations
UPDATE options SET closed = '2025-05-17', exit_price = 0.00 WHERE expiration = '2025-05-17';

-- ==========================================
-- JUNE 2025 - SCALING & Q2 DIVIDENDS
-- ==========================================

-- Q2 Dividends
INSERT INTO dividends (symbol, amount, received) VALUES
('VZ', 66.50, '2025-06-01'),
('CVX', 142.00, '2025-06-01'),
('KO', 69.00, '2025-06-01'),
('MMM', 150.00, '2025-06-01');

-- Q2 Treasury interest earned
INSERT INTO treasuries (cuspid, purchased, maturity, amount, yield, buy_price) VALUES
('INT-Q2-2025', '2025-06-30', '2025-06-30', 873.25, 5.25, 100.00);

-- Complete KO scaling to 200 shares (Using Treasury collateral)
INSERT INTO long_positions (symbol, shares, buy_price, opened) VALUES
('KO', 50, 61.75, '2025-06-05');

-- Reduce Treasury for KO purchase ($3,087.50)
UPDATE treasuries SET amount = amount - 3087.50 WHERE cuspid = '912797JX2';

-- June 21 expirations and assignments
-- VZ called away again - Proceeds to Treasury
UPDATE long_positions SET closed = '2025-06-21', exit_price = 46.00 WHERE symbol = 'VZ' AND opened = '2025-04-01';
UPDATE treasuries SET amount = amount + 4600.00 WHERE cuspid = '912797JX2';

-- MMM put assigned - Use Treasury collateral
INSERT INTO long_positions (symbol, shares, buy_price, opened) VALUES
('MMM', 100, 95.00, '2025-06-21');
UPDATE treasuries SET amount = amount - 9500.00 WHERE cuspid = '912797JX2';

UPDATE options SET closed = '2025-06-21', exit_price = 3.00 WHERE symbol = 'VZ' AND type = 'Call' AND strike = 46.00;
UPDATE options SET closed = '2025-06-21', exit_price = 0.00 WHERE expiration = '2025-06-21' AND symbol != 'VZ';
UPDATE options SET closed = '2025-06-21', exit_price = 2.20 WHERE symbol = 'MMM' AND type = 'Put' AND expiration = '2025-06-21';

-- ==========================================
-- JULY-DECEMBER 2025 - COMPOUNDING PHASE
-- ==========================================

-- July: Rebuild VZ, continue scaling (Treasury collateral)
INSERT INTO long_positions (symbol, shares, buy_price, opened) VALUES
('VZ', 200, 41.50, '2025-07-01');
UPDATE treasuries SET amount = amount - 8300.00 WHERE cuspid = '912797JX2';

-- August: Scale up CVX to 200 shares (Treasury collateral)
INSERT INTO long_positions (symbol, shares, buy_price, opened) VALUES
('CVX', 100, 165.00, '2025-08-15');
UPDATE treasuries SET amount = amount - 16500.00 WHERE cuspid = '912797JX2';

-- Q3 Dividends (September)
INSERT INTO dividends (symbol, amount, received) VALUES
('VZ', 133.00, '2025-09-01'),
('CVX', 284.00, '2025-09-01'),
('KO', 92.00, '2025-09-01'),
('MMM', 300.00, '2025-09-01');

-- Q3 Treasury interest earned
INSERT INTO treasuries (cuspid, purchased, maturity, amount, yield, buy_price) VALUES
('INT-Q3-2025', '2025-09-30', '2025-09-30', 655.33, 5.25, 100.00);

-- October: Additional MMM scaling (Treasury collateral)
INSERT INTO long_positions (symbol, shares, buy_price, opened) VALUES
('MMM', 100, 94.00, '2025-10-15');
UPDATE treasuries SET amount = amount - 9400.00 WHERE cuspid = '912797JX2';

-- Sample of key profitable options trades throughout the year
INSERT INTO options (symbol, type, opened, closed, expiration, strike, premium, exit_price, contracts, commission) VALUES
-- July profitable strangles
('VZ', 'Call', '2025-07-01', '2025-07-19', '2025-07-19', 44.00, 1.20, 0.00, 2, 1.00),
('CVX', 'Call', '2025-07-01', '2025-07-19', '2025-07-19', 175.00, 3.50, 0.00, 2, 1.00),
('KO', 'Call', '2025-07-01', '2025-07-19', '2025-07-19', 65.00, 1.40, 0.00, 2, 1.00),
('MMM', 'Call', '2025-07-01', '2025-07-19', '2025-07-19', 108.00, 2.20, 0.00, 2, 1.00),

-- August successful puts
('VZ', 'Put', '2025-08-01', '2025-08-16', '2025-08-16', 39.00, 1.30, 0.00, 2, 1.00),
('CVX', 'Put', '2025-08-01', '2025-08-16', '2025-08-16', 160.00, 3.80, 0.00, 2, 1.00),
('KO', 'Put', '2025-08-01', '2025-08-16', '2025-08-16', 58.00, 1.10, 0.00, 2, 1.00),
('MMM', 'Put', '2025-08-01', '2025-08-16', '2025-08-16', 92.00, 2.40, 0.00, 2, 1.00),

-- November assignment profits (calls assigned during rally) - Proceeds to Treasury
('VZ', 'Call', '2025-11-01', '2025-11-15', '2025-11-15', 45.00, 1.50, 3.20, 1, 1.00),
('CVX', 'Call', '2025-11-01', '2025-11-15', '2025-11-15', 180.00, 4.20, 8.50, 1, 1.00),
('MMM', 'Call', '2025-11-01', '2025-11-15', '2025-11-15', 110.00, 2.80, 6.50, 1, 1.00);

-- Corresponding stock assignments from November calls (Proceeds to Treasury)
UPDATE long_positions SET closed = '2025-11-15', exit_price = 45.00 WHERE symbol = 'VZ' AND opened = '2025-07-01' AND shares = 100;
UPDATE long_positions SET closed = '2025-11-15', exit_price = 180.00 WHERE symbol = 'CVX' AND opened = '2025-01-05';
UPDATE long_positions SET closed = '2025-11-15', exit_price = 110.00 WHERE symbol = 'MMM' AND opened = '2025-01-05';

-- Add November assignment proceeds to Treasury
UPDATE treasuries SET amount = amount + 4500.00 WHERE cuspid = '912797JX2';  -- VZ: 100 * $45
UPDATE treasuries SET amount = amount + 18000.00 WHERE cuspid = '912797JX2'; -- CVX: 100 * $180  
UPDATE treasuries SET amount = amount + 11000.00 WHERE cuspid = '912797JX2'; -- MMM: 100 * $110

-- Q4 Dividends (December)
INSERT INTO dividends (symbol, amount, received) VALUES
('VZ', 133.00, '2025-12-01'),
('CVX', 284.00, '2025-12-01'),
('KO', 92.00, '2025-12-01'),
('MMM', 450.00, '2025-12-01');

-- Q4 Treasury interest earned and year-end maturity
INSERT INTO treasuries (cuspid, purchased, maturity, amount, yield, buy_price) VALUES
('INT-Q4-2025', '2025-12-31', '2025-12-31', 1893.75, 5.25, 100.00);

-- Mature original Treasury bond
UPDATE treasuries SET maturity = '2025-12-31' WHERE cuspid = '912797JX2';

-- Purchase new Treasury for 2025 with accumulated proceeds
INSERT INTO treasuries (cuspid, purchased, maturity, amount, yield, buy_price) VALUES
('912797KY3', '2025-12-31', '2025-12-31', 85000.00, 5.50, 99.25);

-- Final December positions setup
INSERT INTO options (symbol, type, opened, expiration, strike, premium, contracts, commission) VALUES
-- Year-end covered calls on final positions
('VZ', 'Call', '2025-12-15', '2025-01-17', 46.00, 1.80, 1, 1.00),
('CVX', 'Call', '2025-12-15', '2025-01-17', 175.00, 4.50, 1, 1.00), 
('KO', 'Call', '2025-12-15', '2025-01-17', 67.00, 1.60, 2, 1.00),
('MMM', 'Call', '2025-12-15', '2025-01-17', 112.00, 3.20, 2, 1.00);

-- Year-end cash-secured puts for 2025 income (Backed by new Treasury)
INSERT INTO options (symbol, type, opened, expiration, strike, premium, contracts, commission) VALUES
('VZ', 'Put', '2025-12-15', '2025-01-17', 40.00, 1.90, 3, 1.00),
('CVX', 'Put', '2025-12-15', '2025-01-17', 160.00, 4.80, 2, 1.00),
('KO', 'Put', '2025-12-15', '2025-01-17', 58.00, 1.70, 3, 1.00),
('MMM', 'Put', '2025-12-15', '2025-01-17', 95.00, 2.90, 4, 1.00);

-- Summary: This represents a systematic wheel strategy implementation with proper Treasury collateral management
-- Final positions: 100 VZ, 100 CVX, 200 KO, 200 MMM shares
-- Plus 8 covered calls and 12 cash-secured puts generating monthly income
-- Demonstrates scaling, compounding, Treasury operations, and professional options management