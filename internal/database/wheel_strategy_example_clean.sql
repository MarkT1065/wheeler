-- Clean CVX Wheel Strategy Demonstration Data
-- This demonstrates a realistic wheel strategy focused on CVX (Chevron)
-- Strategy: Use Treasury collateral, sell cash-secured puts, handle assignments, sell covered calls

DELETE FROM options;
DELETE FROM long_positions;
DELETE FROM dividends;
DELETE FROM treasuries;
DELETE FROM symbols;

-- Step 1: Initialize all symbols with current market data
INSERT INTO symbols (symbol, price, dividend, ex_dividend_date, pe_ratio) VALUES
('CVX', 161.45, 6.84, '2025-02-12', 15.2),
('KO', 62.85, 2.00, '2025-03-15', 26.1),
('VZ', 41.20, 2.71, '2025-02-07', 9.8),
('MRK', 98.75, 3.22, '2025-03-07', 16.7);

-- Step 2: Initial Treasury collateral - $200,000 total investment
INSERT INTO treasuries (cuspid, purchased, maturity, amount, yield, buy_price, current_value) VALUES
('912797WH8', '2025-01-02', '2025-12-15', 100000.0, 4.5, 99750.00, 100000.0),
('912797WH9', '2025-01-02', '2025-06-15', 50000.0, 4.4, 49850.00, 50000.0),
('912797WI0', '2025-01-02', '2026-01-15', 50000.0, 4.6, 49725.00, 50000.0);

-- WEEK 1 (Jan 3-9, 2025): Start diversified wheel strategy
-- CVX at $158, KO at $61.50, VZ at $40.80, MRK at $97.25
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
-- CVX positions (25% allocation = $50K)
('CVX', 'Put', '2025-01-03', 150.0, '2025-01-17', 0.85, 3, 1.95),  -- $255 premium, $45K exposure
('CVX', 'Put', '2025-01-06', 155.0, '2025-01-17', 1.95, 2, 1.30),  -- $390 premium, $31K exposure
-- KO positions (25% allocation = $50K) 
('KO', 'Put', '2025-01-03', 60.0, '2025-01-17', 1.15, 8, 5.20),   -- $920 premium, $48K exposure
('KO', 'Put', '2025-01-07', 58.0, '2025-01-24', 0.85, 2, 1.30),   -- $170 premium, $11.6K exposure
-- VZ positions (25% allocation = $50K)
('VZ', 'Put', '2025-01-03', 39.0, '2025-01-17', 0.65, 12, 7.80),  -- $780 premium, $46.8K exposure
('VZ', 'Put', '2025-01-08', 40.0, '2025-01-24', 0.95, 3, 1.95),   -- $285 premium, $12K exposure
-- MRK positions (25% allocation = $50K)
('MRK', 'Put', '2025-01-03', 95.0, '2025-01-17', 1.45, 5, 3.25),  -- $725 premium, $47.5K exposure
('MRK', 'Put', '2025-01-06', 92.0, '2025-01-24', 0.95, 1, 0.65);  -- $95 premium, $9.2K exposure

-- WEEK 2 (Jan 10-16): Continue diversified selling as markets dip
-- CVX at $154, KO at $60.85, VZ at $40.25, MRK at $96.50
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
-- CVX additional positions
('CVX', 'Put', '2025-01-10', 145.0, '2025-01-24', 0.75, 2, 1.30),  -- $150 premium, $29K exposure
('CVX', 'Put', '2025-01-15', 148.0, '2025-01-31', 1.05, 2, 1.30),  -- $210 premium, $29.6K exposure
-- KO additional positions  
('KO', 'Put', '2025-01-10', 58.0, '2025-01-31', 0.95, 4, 2.60),   -- $380 premium, $23.2K exposure
('KO', 'Put', '2025-01-14', 59.0, '2025-02-07', 1.25, 3, 1.95),   -- $375 premium, $17.7K exposure
-- VZ additional positions
('VZ', 'Put', '2025-01-10', 38.0, '2025-01-31', 0.55, 6, 3.90),   -- $330 premium, $22.8K exposure
('VZ', 'Put', '2025-01-13', 39.5, '2025-02-07', 0.85, 4, 2.60),   -- $340 premium, $15.8K exposure
-- MRK additional positions
('MRK', 'Put', '2025-01-10', 93.0, '2025-01-31', 1.15, 3, 1.95),  -- $345 premium, $27.9K exposure
('MRK', 'Put', '2025-01-15', 90.0, '2025-02-07', 0.85, 2, 1.30);  -- $170 premium, $18K exposure

-- WEEK 3 (Jan 17): First multi-symbol expiration - most puts expire worthless
-- CVX rallies to $156, KO to $61.80, VZ stable at $40.50, MRK to $98.10
UPDATE options SET closed = '2025-01-17', exit_price = 0.05 
WHERE strike = 150.0 AND expiration = '2025-01-17' AND type = 'Put' AND symbol = 'CVX';
UPDATE options SET closed = '2025-01-17', exit_price = 0.10 
WHERE strike = 155.0 AND expiration = '2025-01-17' AND type = 'Put' AND symbol = 'CVX';
UPDATE options SET closed = '2025-01-17', exit_price = 0.08 
WHERE strike = 60.0 AND expiration = '2025-01-17' AND type = 'Put' AND symbol = 'KO';
UPDATE options SET closed = '2025-01-17', exit_price = 0.12 
WHERE strike = 39.0 AND expiration = '2025-01-17' AND type = 'Put' AND symbol = 'VZ';
UPDATE options SET closed = '2025-01-17', exit_price = 0.15 
WHERE strike = 95.0 AND expiration = '2025-01-17' AND type = 'Put' AND symbol = 'MRK';

-- Continue diversified selling after successful expirations
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
-- CVX follow-up positions
('CVX', 'Put', '2025-01-17', 152.0, '2025-02-07', 1.35, 2, 1.30),  -- $270 premium, $30.4K exposure
('CVX', 'Put', '2025-01-22', 150.0, '2025-02-14', 1.25, 2, 1.30),  -- $250 premium, $30K exposure
-- KO follow-up positions
('KO', 'Put', '2025-01-17', 59.0, '2025-02-14', 1.05, 5, 3.25),   -- $525 premium, $29.5K exposure
('KO', 'Put', '2025-01-21', 60.5, '2025-02-21', 1.35, 3, 1.95),   -- $405 premium, $18.15K exposure
-- VZ follow-up positions
('VZ', 'Put', '2025-01-17', 39.5, '2025-02-14', 0.75, 8, 5.20),   -- $600 premium, $31.6K exposure
('VZ', 'Put', '2025-01-24', 40.5, '2025-02-21', 0.95, 3, 1.95),   -- $285 premium, $12.15K exposure
-- MRK follow-up positions
('MRK', 'Put', '2025-01-17', 92.0, '2025-02-14', 1.25, 4, 2.60),  -- $500 premium, $36.8K exposure
('MRK', 'Put', '2025-01-20', 94.0, '2025-02-21', 1.55, 2, 1.30);  -- $310 premium, $18.8K exposure

-- WEEK 4 (Jan 24): MIXED RESULTS - Diversification shows its value
-- Market selloff: CVX drops to $149, KO stable at $60.75, VZ dips to $39.95, MRK down to $95.80
-- CVX: Some puts expire worthless, some get assigned
UPDATE options SET closed = '2025-01-24', exit_price = 0.12 
WHERE strike = 145.0 AND expiration = '2025-01-24' AND type = 'Put' AND symbol = 'CVX';
-- KO puts expire worthless (stock held above strikes)
UPDATE options SET closed = '2025-01-24', exit_price = 0.08 
WHERE strike = 58.0 AND expiration = '2025-01-24' AND type = 'Put' AND symbol = 'KO';
-- VZ gets assignment at $40 strike (stock at $39.95)
UPDATE options SET closed = '2025-01-24', exit_price = 0.05 
WHERE strike = 40.0 AND expiration = '2025-01-24' AND type = 'Put' AND symbol = 'VZ';
-- MRK puts expire worthless
UPDATE options SET closed = '2025-01-24', exit_price = 0.15 
WHERE strike = 92.0 AND expiration = '2025-01-24' AND type = 'Put' AND symbol = 'MRK';

-- ASSIGNMENTS: VZ gets assigned (3 contracts x 100 shares at $40)
INSERT INTO long_positions (symbol, opened, shares, buy_price) VALUES
('VZ', '2025-01-24', 300, 40.0);

-- Adjust treasuries for VZ stock purchase ($12,000)
UPDATE treasuries SET 
    current_value = 38000.0,
    amount = 38000.0
WHERE cuspid = '912797WH9';

-- Start selling COVERED CALLS on VZ assignment and continue puts on non-assigned symbols
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
('VZ', 'Call', '2025-01-27', 42.0, '2025-02-14', 1.25, 3, 1.95),  -- $375 premium on VZ stock
('CVX', 'Put', '2025-01-27', 145.0, '2025-02-14', 1.15, 3, 1.95),  -- $345 premium, $43.5K exposure
('KO', 'Put', '2025-01-27', 58.5, '2025-02-14', 1.05, 4, 2.60),   -- $420 premium, $23.4K exposure
('MRK', 'Put', '2025-01-27', 93.0, '2025-02-14', 1.35, 3, 1.95);  -- $405 premium, $27.9K exposure

-- WEEK 5 (Jan 31): More puts expire worthless
UPDATE options SET closed = '2025-01-31', exit_price = 0.12 
WHERE strike = 148.0 AND expiration = '2025-01-31' AND type = 'Put';
UPDATE options SET closed = '2025-01-31', exit_price = 0.15 
WHERE strike = 150.0 AND expiration = '2025-01-31' AND type = 'Put';

-- Continue selling puts with remaining collateral
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
('CVX', 'Put', '2025-02-03', 145.0, '2025-02-21', 0.85, 3, 1.95),  -- $255 premium, commission: 3 × $0.65
('CVX', 'Put', '2025-02-05', 148.0, '2025-02-21', 1.25, 2, 1.30);  -- $250 premium, commission: 2 × $0.65

-- WEEK 6 (Feb 7): Fourth expiration - puts expire worthless
UPDATE options SET closed = '2025-02-07', exit_price = 0.10 
WHERE strike = 152.0 AND expiration = '2025-02-07' AND type = 'Put';
UPDATE options SET closed = '2025-02-07', exit_price = 0.08 
WHERE strike = 147.0 AND expiration = '2025-02-07' AND type = 'Put';

-- WEEK 7 (Feb 14): CALLS EXERCISED - stock rallied to $157
-- Sell 300 shares at $155 strike (profitable wheel completion)
UPDATE options SET closed = '2025-02-14', exit_price = 0.05 
WHERE type = 'Call' AND strike = 155.0 AND expiration = '2025-02-14';
UPDATE long_positions SET 
    closed = '2025-02-14',
    exit_price = 155.0 
WHERE symbol = 'CVX' AND opened = '2025-01-24';

-- Dividend received while holding stock
INSERT INTO dividends (symbol, received, amount) VALUES
('CVX', '2025-02-12', 513.0);  -- 300 shares * $1.71 quarterly dividend

-- Buy back treasuries with stock sale proceeds
INSERT INTO treasuries (cuspid, purchased, maturity, amount, yield, buy_price, current_value) VALUES
('912797WI6', '2025-02-14', '2026-01-15', 46500.0, 4.3, 46275.00, 46500.0);

-- WEEK 8-9: Resume selling puts after successful wheel cycle
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
('CVX', 'Put', '2025-02-18', 150.0, '2025-03-07', 1.45, 4, 2.60),  -- $580 premium, 4 × $0.65 commission
('CVX', 'Put', '2025-02-20', 155.0, '2025-03-07', 2.25, 2, 1.30),  -- $450 premium, 2 × $0.65 commission
('CVX', 'Put', '2025-02-21', 147.0, '2025-02-28', 0.95, 2, 1.30);  -- $190 premium, 2 × $0.65 commission

-- WEEK 9 (Feb 21): Puts expire worthless
UPDATE options SET closed = '2025-02-21', exit_price = 0.08 
WHERE strike = 145.0 AND expiration = '2025-02-21' AND type = 'Put';
UPDATE options SET closed = '2025-02-21', exit_price = 0.12 
WHERE strike = 148.0 AND expiration = '2025-02-21' AND type = 'Put';

-- WEEK 10 (Feb 28): More profitable expirations
UPDATE options SET closed = '2025-02-28', exit_price = 0.10 
WHERE strike = 147.0 AND expiration = '2025-02-28' AND type = 'Put';

-- WEEK 11 (Mar 7): SECOND ASSIGNMENT CYCLE - market dropped to $148
UPDATE options SET closed = '2025-03-07', exit_price = 0.15 
WHERE strike = 150.0 AND expiration = '2025-03-07' AND type = 'Put';

-- SECOND ASSIGNMENT: Buy 200 shares at $155
INSERT INTO long_positions (symbol, opened, shares, buy_price) VALUES
('CVX', '2025-03-07', 200, 155.0);

-- Adjust treasuries for stock purchase ($31,000)
UPDATE treasuries SET 
    current_value = 15500.0,
    amount = 15500.0 
WHERE cuspid = '912797WI6';

-- Start covered calls on new stock position
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
('CVX', 'Call', '2025-03-10', 160.0, '2025-03-21', 1.85, 2, 1.30);  -- $370 premium, 2 × $0.65 commission

-- Continue selling puts with remaining collateral
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
('CVX', 'Put', '2025-03-12', 145.0, '2025-03-28', 1.15, 1, 0.65),  -- $115 premium, 1 × $0.65 commission
('CVX', 'Put', '2025-03-14', 150.0, '2025-03-28', 1.65, 1, 0.65);  -- $165 premium, 1 × $0.65 commission

-- WEEK 13 (Mar 21): Calls expire worthless, roll forward
UPDATE options SET closed = '2025-03-21', exit_price = 0.05 
WHERE type = 'Call' AND strike = 160.0 AND expiration = '2025-03-21';

-- Roll calls to higher strike
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
('CVX', 'Call', '2025-03-24', 158.0, '2025-04-11', 2.25, 2, 1.30);  -- $450 premium, 2 × $0.65 commission

-- WEEK 14 (Mar 28): Puts expire worthless
UPDATE options SET closed = '2025-03-28', exit_price = 0.12 
WHERE strike = 145.0 AND expiration = '2025-03-28' AND type = 'Put';
UPDATE options SET closed = '2025-03-28', exit_price = 0.18 
WHERE strike = 150.0 AND expiration = '2025-03-28' AND type = 'Put';

-- Continue selling puts
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
('CVX', 'Put', '2025-03-31', 148.0, '2025-04-18', 1.35, 2, 1.30),  -- $270 premium, 2 × $0.65 commission
('CVX', 'Put', '2025-04-02', 152.0, '2025-04-18', 1.85, 1, 0.65);  -- $185 premium, 1 × $0.65 commission

-- Add recent profitable trades
INSERT INTO options (symbol, type, opened, closed, strike, expiration, premium, contracts, exit_price, commission) VALUES
('CVX', 'Put', '2025-03-17', '2025-04-04', 145.0, '2025-04-04', 0.85, 3, 0.08, 3.50),
('CVX', 'Put', '2025-03-20', '2025-04-04', 148.0, '2025-04-04', 1.15, 2, 0.12, 2.50),
('CVX', 'Put', '2025-03-25', '2025-04-04', 150.0, '2025-04-04', 1.45, 2, 0.15, 2.50);

-- Add quarterly dividend
INSERT INTO dividends (symbol, received, amount) VALUES
('CVX', '2025-03-15', 342.0);  -- 200 shares * $1.71

-- WEEK 15-16 (April): Continue wheel strategy
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
('CVX', 'Put', '2025-04-07', 150.0, '2025-04-25', 1.25, 2, 1.30),  -- $250 premium, 2 × $0.65 commission
('CVX', 'Put', '2025-04-09', 155.0, '2025-04-25', 1.95, 1, 0.65),  -- $195 premium, 1 × $0.65 commission
('CVX', 'Call', '2025-04-10', 165.0, '2025-04-25', 1.75, 2, 1.30); -- $350 premium, 2 × $0.65 commission

-- WEEK 17 (Apr 25): Calls expire worthless, puts assigned again at $155
UPDATE options SET closed = '2025-04-25', exit_price = 0.08 
WHERE expiration = '2025-04-25' AND type = 'Call';
UPDATE options SET closed = '2025-04-25', exit_price = 0.12 
WHERE strike = 150.0 AND expiration = '2025-04-25' AND type = 'Put';

-- THIRD ASSIGNMENT: 100 more shares at $155 (now have 300 total)
INSERT INTO long_positions (symbol, opened, shares, buy_price) VALUES
('CVX', '2025-04-25', 100, 155.0);

-- Adjust treasuries for additional stock purchase
UPDATE treasuries SET 
    current_value = 0.0,
    amount = 0.0 
WHERE cuspid = '912797WI6';

-- MAY: Sell covered calls on expanded 300-share position
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
('CVX', 'Call', '2025-05-02', 162.0, '2025-05-16', 2.45, 3, 1.95),  -- $735 premium, 3 × $0.65 commission
('CVX', 'Put', '2025-05-05', 148.0, '2025-05-30', 1.35, 2, 1.30);   -- $270 premium, 2 × $0.65 commission

-- Third quarterly dividend
INSERT INTO dividends (symbol, received, amount) VALUES
('CVX', '2025-05-15', 513.0);  -- 300 shares * $1.71

-- MAY 16: Calls expire worthless, continue wheel
UPDATE options SET closed = '2025-05-16', exit_price = 0.10 
WHERE type = 'Call' AND expiration = '2025-05-16';

-- JUNE: Roll forward calls and continue puts
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
('CVX', 'Call', '2025-06-03', 165.0, '2025-06-20', 2.85, 3, 1.95),  -- $855 premium, 3 × $0.65 commission
('CVX', 'Put', '2025-06-10', 150.0, '2025-06-20', 1.55, 2, 1.30),   -- $310 premium, 2 × $0.65 commission
('CVX', 'Put', '2025-06-12', 152.0, '2025-06-27', 1.75, 1, 0.65);   -- $175 premium, 1 × $0.65 commission

-- JUNE 20: Market rallies to $167 - CALLS EXERCISED! 
UPDATE options SET closed = '2025-06-20', exit_price = 0.05 
WHERE type = 'Call' AND expiration = '2025-06-20';
UPDATE options SET closed = '2025-06-20', exit_price = 0.12 
WHERE expiration = '2025-06-20' AND type = 'Put';

-- Sell all 300 shares at $165 (major profit realization)
UPDATE long_positions SET 
    closed = '2025-06-20',
    exit_price = 165.0 
WHERE symbol = 'CVX' AND closed IS NULL;

-- Replenish treasury collateral with stock sale proceeds
INSERT INTO treasuries (cuspid, purchased, maturity, amount, yield, buy_price, current_value) VALUES
('912797WJ4', '2025-06-20', '2026-06-15', 49500.0, 4.6, 49225.00, 49500.0);

-- JULY-AUGUST: Continue diversified strategies across all symbols
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
-- CVX positions (energy sector)
('CVX', 'Put', '2025-07-01', 155.0, '2025-07-18', 2.15, 3, 1.95),   -- $645 premium
('CVX', 'Put', '2025-07-08', 160.0, '2025-07-25', 2.85, 2, 1.30),   -- $570 premium
('CVX', 'Put', '2025-07-15', 158.0, '2025-08-15', 2.45, 3, 1.95),   -- $735 premium
('CVX', 'Put', '2025-08-01', 152.0, '2025-08-15', 1.95, 2, 1.30),   -- $390 premium
('CVX', 'Put', '2025-08-12', 155.0, '2025-08-29', 2.25, 2, 1.30),   -- $450 premium
-- KO positions (consumer staples)
('KO', 'Put', '2025-07-01', 62.0, '2025-07-18', 1.85, 6, 3.90),     -- $1110 premium, $37.2K exposure
('KO', 'Put', '2025-07-10', 60.0, '2025-07-25', 1.45, 4, 2.60),     -- $580 premium, $24K exposure  
('KO', 'Put', '2025-08-01', 58.5, '2025-08-15', 1.25, 5, 3.25),     -- $625 premium, $29.25K exposure
('KO', 'Put', '2025-08-15', 61.0, '2025-08-29', 1.75, 3, 1.95),     -- $525 premium, $18.3K exposure
-- VZ positions (telecom)
('VZ', 'Put', '2025-07-05', 41.0, '2025-07-18', 0.95, 8, 5.20),     -- $760 premium, $32.8K exposure
('VZ', 'Put', '2025-07-15', 39.5, '2025-08-15', 0.75, 6, 3.90),     -- $450 premium, $23.7K exposure
('VZ', 'Put', '2025-08-05', 40.5, '2025-08-29', 0.85, 7, 4.55),     -- $595 premium, $28.35K exposure
-- MRK positions (healthcare) 
('MRK', 'Put', '2025-07-02', 96.0, '2025-07-18', 2.25, 4, 2.60),    -- $900 premium, $38.4K exposure
('MRK', 'Put', '2025-07-12', 94.0, '2025-08-15', 1.85, 3, 1.95),    -- $555 premium, $28.2K exposure  
('MRK', 'Put', '2025-08-08', 98.0, '2025-08-29', 2.45, 2, 1.30);    -- $490 premium, $19.6K exposure

-- Most July/August puts expire worthless across all symbols (market strength)
UPDATE options SET closed = '2025-07-18', exit_price = 0.15 
WHERE expiration = '2025-07-18' AND type = 'Put';
UPDATE options SET closed = '2025-07-25', exit_price = 0.22 
WHERE expiration = '2025-07-25' AND type = 'Put';
UPDATE options SET closed = '2025-08-15', exit_price = 0.18 
WHERE expiration = '2025-08-15' AND type = 'Put';
UPDATE options SET closed = '2025-08-29', exit_price = 0.12 
WHERE expiration = '2025-08-29' AND type = 'Put';

-- Fourth quarterly dividend (we don't own stock, so no dividend this quarter)

-- SEPTEMBER: Market volatility increases across all symbols
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
-- CVX positions (energy volatility)
('CVX', 'Put', '2025-09-02', 158.0, '2025-09-20', 2.95, 3, 1.95),   -- $885 premium
('CVX', 'Put', '2025-09-10', 162.0, '2025-09-20', 3.45, 2, 1.30),   -- $690 premium
('CVX', 'Put', '2025-09-15', 155.0, '2025-10-17', 2.15, 3, 1.95),   -- $645 premium
-- KO positions (defensive plays during volatility)
('KO', 'Put', '2025-09-03', 61.5, '2025-09-20', 1.95, 4, 2.60),     -- $780 premium, $24.6K exposure
('KO', 'Put', '2025-09-12', 63.0, '2025-10-17', 2.25, 3, 1.95),     -- $675 premium, $18.9K exposure
('KO', 'Put', '2025-09-18', 59.0, '2025-10-03', 1.35, 5, 3.25),     -- $675 premium, $29.5K exposure
-- VZ positions (telecom stability)
('VZ', 'Put', '2025-09-05', 42.0, '2025-09-20', 1.15, 6, 3.90),     -- $690 premium, $25.2K exposure  
('VZ', 'Put', '2025-09-15', 40.0, '2025-10-17', 0.95, 8, 5.20),     -- $760 premium, $32K exposure
('VZ', 'Put', '2025-09-25', 43.0, '2025-10-31', 1.35, 4, 2.60),     -- $540 premium, $17.2K exposure
-- MRK positions (healthcare strength)
('MRK', 'Put', '2025-09-01', 99.0, '2025-09-20', 2.85, 3, 1.95),    -- $855 premium, $29.7K exposure
('MRK', 'Put', '2025-09-10', 101.0, '2025-10-17', 3.25, 2, 1.30),   -- $650 premium, $20.2K exposure
('MRK', 'Put', '2025-09-20', 97.0, '2025-10-31', 2.45, 4, 2.60);    -- $980 premium, $38.8K exposure

-- SEPTEMBER 20: Mixed assignment results across symbols
-- CVX drops to $159, KO stable at $62.80, VZ dips to $41.75, MRK at $100.50
UPDATE options SET closed = '2025-09-20', exit_price = 0.25 
WHERE expiration = '2025-09-20' AND type = 'Put';

-- ASSIGNMENTS: VZ gets assigned (6 contracts at $42), others expire worthless
INSERT INTO long_positions (symbol, opened, shares, buy_price) VALUES
('VZ', '2025-09-20', 600, 42.0);  -- VZ assignment: 6 × 100 shares

-- Adjust treasury for VZ purchase ($25,200)
UPDATE treasuries SET 
    current_value = 26300.0,
    amount = 26300.0 
WHERE cuspid = '912797WJ4';

-- OCTOBER: VZ covered calls + continue puts on other symbols
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
-- VZ covered calls (600 shares assigned)
('VZ', 'Call', '2025-10-01', 44.0, '2025-10-18', 1.45, 6, 3.90),    -- $870 premium on VZ stock
-- CVX puts (no assignment)
('CVX', 'Put', '2025-10-03', 155.0, '2025-10-18', 2.25, 3, 1.95),   -- $675 premium, $46.5K exposure
('CVX', 'Put', '2025-10-10', 158.0, '2025-11-15', 2.85, 2, 1.30),   -- $570 premium, $31.6K exposure  
-- KO puts continue
('KO', 'Put', '2025-10-01', 62.5, '2025-10-18', 2.05, 4, 2.60),     -- $820 premium, $25K exposure
('KO', 'Put', '2025-10-08', 64.0, '2025-11-15', 2.45, 3, 1.95),     -- $735 premium, $19.2K exposure
-- MRK puts continue  
('MRK', 'Put', '2025-10-05', 100.0, '2025-10-18', 3.15, 3, 1.95),   -- $945 premium, $30K exposure
('MRK', 'Put', '2025-10-12', 102.0, '2025-11-15', 3.65, 2, 1.30);   -- $730 premium, $20.4K exposure

-- October 18 expirations
UPDATE options SET closed = '2025-10-18', exit_price = 0.12 
WHERE expiration = '2025-10-18';

-- October 3 expirations  
UPDATE options SET closed = '2025-10-03', exit_price = 0.08
WHERE expiration = '2025-10-03';

-- NOVEMBER: Multi-symbol final quarter push
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
-- VZ covered calls (600 shares)
('VZ', 'Call', '2025-11-01', 45.0, '2025-11-15', 1.75, 6, 3.90),    -- $1050 premium
-- CVX puts continue
('CVX', 'Put', '2025-11-02', 160.0, '2025-11-15', 2.95, 3, 1.95),   -- $885 premium, $48K exposure
('CVX', 'Put', '2025-11-10', 155.0, '2025-11-29', 2.25, 2, 1.30),   -- $450 premium, $31K exposure
-- KO assignment scenario (gets assigned)
('KO', 'Put', '2025-11-01', 63.5, '2025-11-15', 2.35, 4, 2.60),     -- $940 premium, $25.4K exposure  
-- MRK puts continue
('MRK', 'Put', '2025-11-05', 101.5, '2025-11-29', 3.45, 3, 1.95),   -- $1035 premium, $30.45K exposure
('MRK', 'Put', '2025-11-12', 99.0, '2025-12-20', 2.85, 2, 1.30);    -- $570 premium, $19.8K exposure

-- NOVEMBER 15: Mixed results across symbols
-- VZ calls exercised (stock rallies to $46.50), KO gets assigned (drops to $62.80)
UPDATE options SET closed = '2025-11-15', exit_price = 0.08 
WHERE expiration = '2025-11-15';

-- VZ calls exercised - sell 600 shares at $45
UPDATE long_positions SET 
    closed = '2025-11-15',
    exit_price = 45.0 
WHERE symbol = 'VZ' AND opened = '2025-09-20';

-- KO assignment - buy 400 shares at $63.50
INSERT INTO long_positions (symbol, opened, shares, buy_price) VALUES
('KO', '2025-11-15', 400, 63.5);

-- Adjust treasuries (VZ sale +$27K, KO purchase -$25.4K, net +$1.6K)
UPDATE treasuries SET 
    current_value = 27900.0,
    amount = 27900.0 
WHERE cuspid = '912797WJ4';

-- DECEMBER: Year-end positioning across all symbols
INSERT INTO options (symbol, type, opened, strike, expiration, premium, contracts, commission) VALUES
-- CVX year-end puts
('CVX', 'Put', '2025-12-01', 160.0, '2025-12-20', 3.25, 3, 1.95),   -- $975 premium
('CVX', 'Put', '2025-12-10', 165.0, '2025-12-20', 4.15, 2, 1.30),   -- $830 premium
('CVX', 'Put', '2025-12-15', 158.0, '2026-01-17', 2.85, 2, 1.30),   -- $570 premium
-- KO covered calls (400 shares assigned)
('KO', 'Call', '2025-12-01', 66.0, '2025-12-20', 2.25, 4, 2.60),    -- $900 premium
('KO', 'Put', '2025-12-10', 61.0, '2026-01-17', 1.95, 3, 1.95),     -- $585 premium, $18.3K exposure
-- VZ back to puts (no stock after November call exercise)
('VZ', 'Put', '2025-12-02', 42.5, '2025-12-20', 1.45, 6, 3.90),     -- $870 premium, $25.5K exposure
('VZ', 'Put', '2025-12-12', 44.0, '2026-01-17', 1.75, 4, 2.60),     -- $700 premium, $17.6K exposure
-- MRK final pushes
('MRK', 'Put', '2025-12-05', 103.0, '2025-12-20', 3.85, 3, 1.95),   -- $1155 premium, $30.9K exposure
('MRK', 'Put', '2025-12-15', 100.5, '2026-01-17', 3.25, 2, 1.30);   -- $650 premium, $20.1K exposure

-- December 20: Most expire worthless, KO calls exercised
UPDATE options SET closed = '2025-12-20', exit_price = 0.18 
WHERE expiration = '2025-12-20' AND type = 'Put';
UPDATE options SET closed = '2025-12-20', exit_price = 0.08 
WHERE expiration = '2025-12-20' AND type = 'Call';

-- KO calls exercised - sell 400 shares at $66
UPDATE long_positions SET 
    closed = '2025-12-20',
    exit_price = 66.0 
WHERE symbol = 'KO' AND opened = '2025-11-15';

-- Replenish treasuries with KO sale proceeds (+$26.4K)
UPDATE treasuries SET 
    current_value = 54300.0,
    amount = 54300.0 
WHERE cuspid = '912797WJ4';

-- Add realistic quarterly dividends based on actual holdings
INSERT INTO dividends (symbol, received, amount) VALUES 
-- CVX dividends (no holdings most of the year - energy stock)
-- KO dividends (based on actual assignments)
('KO', '2025-12-15', 200.0),    -- Q4: 400 shares × $0.50 (Nov 15 - Dec 20)
-- VZ dividends (based on actual assignments)  
('VZ', '2025-02-07', 203.25),   -- Q1: 300 shares × $0.6775 (Jan 24 - Feb 14)
('VZ', '2025-11-07', 406.50),   -- Q4: 600 shares × $0.6775 (Sep 20 - Nov 15)
-- MRK dividends (no major assignments during dividend periods)
('MRK', '2025-01-15', 80.33);   -- Q4 2024 carryover: 100 shares × $0.8033

-- Update final stock prices for year-end (showing good performance across all)
UPDATE symbols SET price = 172.35, updated_at = CURRENT_TIMESTAMP WHERE symbol = 'CVX';
UPDATE symbols SET price = 65.20, updated_at = CURRENT_TIMESTAMP WHERE symbol = 'KO';  
UPDATE symbols SET price = 43.15, updated_at = CURRENT_TIMESTAMP WHERE symbol = 'VZ';
UPDATE symbols SET price = 102.90, updated_at = CURRENT_TIMESTAMP WHERE symbol = 'MRK';