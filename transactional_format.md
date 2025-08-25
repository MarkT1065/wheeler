# Universal Transaction CSV Format Proposal

## Overview

This document proposes a unified transactional data format for Wheeler that replaces the current multi-CSV approach with a single, granular transaction-based format suitable for precise testing scenarios and complex financial strategy modeling.

## Current Wheeler Import System

Wheeler currently uses three separate CSV formats:

### 1. Options CSV (10 columns)
```csv
symbol,opened,closed,type,strike,expiration,premium,contracts,exit_price,commission
AAPL,2024-01-15,2024-02-20,Put,150.00,2024-03-15,5.50,2,3.25,2.10
```
- **Two-transaction model**: `opened` date creates position, optional `closed`/`exit_price` closes it
- Business logic: Cash-secured puts and covered calls

### 2. Stocks CSV (6 columns) 
```csv
symbol,purchased,closed_date,shares,buy_price,exit_price
AAPL,1/15/2024,2/20/2024,200,145.50,152.75
```
- **Two-transaction model**: `purchased` opens position, optional `closed_date`/`exit_price` closes it
- Note: `shares` field is decimal representing hundreds (2.0 = 200 shares)

### 3. Dividends CSV (3 columns)
```csv
symbol,date_received,amount
AAPL,1/15/2024,25.50
```
- **Single-transaction model**: Just records dividend payment

## Proposed Unified Transactional Format

### Universal Transaction CSV Schema

```csv
transaction_type,symbol,date,action,quantity,price,strike,expiration,option_type,amount,commission,notes
```

### Field Definitions

| Field | Type | Description | Required For |
|-------|------|-------------|--------------|
| `transaction_type` | String | `STOCK`, `OPTION`, `DIVIDEND` | All |
| `symbol` | String | Stock ticker symbol | All |
| `date` | Date | Transaction date (YYYY-MM-DD) | All |
| `action` | String | Transaction action (see actions below) | All |
| `quantity` | Integer | Number of shares or contracts | Stock, Options |
| `price` | Decimal | Price per share or option premium | Stock, Options |
| `strike` | Decimal | Strike price | Options only |
| `expiration` | Date | Expiration date (YYYY-MM-DD) | Options only |
| `option_type` | String | `Put` or `Call` | Options only |
| `amount` | Decimal | Direct monetary amount | Dividends, fees |
| `commission` | Decimal | Transaction commission/fees | Optional |
| `notes` | String | Free-form description | Optional |

### Supported Actions

#### Stock Actions
- `BUY` - Purchase shares
- `SELL` - Sell shares

#### Option Actions
- `SELL_TO_OPEN` - Sell option to open position (collect premium)
- `BUY_TO_CLOSE` - Buy option to close position (pay premium)
- `ASSIGNED` - Option assignment (automatic stock purchase/sale)
- `EXPIRED` - Option expired worthless

#### Dividend Actions
- `RECEIVE` - Dividend payment received

## Example Transactions

### Basic Stock Trade
```csv
transaction_type,symbol,date,action,quantity,price,strike,expiration,option_type,amount,commission,notes
STOCK,AAPL,2024-01-15,BUY,200,145.50,,,,,2.50,Opening position
STOCK,AAPL,2024-02-20,SELL,200,152.75,,,,,2.50,Closing position
```

### Options Trading
```csv
transaction_type,symbol,date,action,quantity,price,strike,expiration,option_type,amount,commission,notes
OPTION,AAPL,2024-01-15,SELL_TO_OPEN,2,5.50,150.00,2024-03-15,Put,,2.10,Cash-secured put
OPTION,AAPL,2024-02-20,BUY_TO_CLOSE,2,3.25,150.00,2024-03-15,Put,,2.10,Closing put early
```

### Dividend Payment
```csv
transaction_type,symbol,date,action,quantity,price,strike,expiration,option_type,amount,commission,notes
DIVIDEND,AAPL,2024-01-25,RECEIVE,,,,,,,25.50,Quarterly dividend
```

### Complex Wheel Strategy Scenario
```csv
transaction_type,symbol,date,action,quantity,price,strike,expiration,option_type,amount,commission,notes
OPTION,AAPL,2024-01-15,SELL_TO_OPEN,1,5.50,150.00,2024-02-16,Put,,1.05,Open cash-secured put
STOCK,AAPL,2024-02-16,BUY,100,150.00,,,,,0,Put assignment (stock purchased)
OPTION,AAPL,2024-02-16,ASSIGNED,1,0,150.00,2024-02-16,Put,,0,Put assigned
OPTION,AAPL,2024-02-17,SELL_TO_OPEN,1,3.25,155.00,2024-03-15,Call,,1.05,Covered call on assigned stock
DIVIDEND,AAPL,2024-03-01,RECEIVE,,,,,,,18.50,Dividend while holding stock
OPTION,AAPL,2024-03-15,EXPIRED,1,0,155.00,2024-03-15,Call,,0,Call expired worthless
```

## Benefits

### 1. Granular Testing
- Each transaction can be tested individually
- Precise control over test scenarios and assertions
- Easy to verify specific business logic rules

### 2. Flexible Scenarios
- Model partial closes and complex strategies
- Handle option assignments and early exercises
- Support unusual market events

### 3. Complete Audit Trail
- Full transaction history for compliance
- Easy to reconstruct position states at any point in time
- Clear chain of causality between related transactions

### 4. Extensible Design
- Easy to add new transaction types (stock splits, spinoffs, etc.)
- Future-proof for additional trading strategies
- Maintains backward compatibility through conversion utilities

## Test Scenario Applications

### Wheel Strategy Testing
```csv
# Test: Put assignment followed by covered call
OPTION,AAPL,2024-01-15,SELL_TO_OPEN,1,5.50,150.00,2024-02-16,Put,,1.05,
STOCK,AAPL,2024-02-16,BUY,100,150.00,,,,,0,Assignment
OPTION,AAPL,2024-02-16,ASSIGNED,1,0,150.00,2024-02-16,Put,,0,
OPTION,AAPL,2024-02-17,SELL_TO_OPEN,1,3.25,155.00,2024-03-15,Call,,1.05,
```

**Assertions:**
- Put premium collected: $550
- Stock position created: 100 shares @ $150
- Call premium collected: $325
- Total premium: $875
- Break-even: $141.25/share

### Multiple Symbol Portfolio
```csv
# Test: Diversified wheel strategy across multiple symbols
OPTION,AAPL,2024-01-15,SELL_TO_OPEN,2,5.50,150.00,2024-03-15,Put,,2.10,
OPTION,MSFT,2024-01-15,SELL_TO_OPEN,1,8.75,300.00,2024-03-15,Put,,1.05,
OPTION,GOOGL,2024-01-15,SELL_TO_OPEN,1,12.25,120.00,2024-03-15,Put,,1.05,
```

**Assertions:**
- Total put exposure: $77,000
- Total premium collected: $2,650
- Portfolio put ROI: 3.44%

## Implementation Notes

### Conversion from Existing Formats
The current three CSV formats can be automatically converted to the universal format:

1. **Options**: Split into SELL_TO_OPEN + optional BUY_TO_CLOSE transactions
2. **Stocks**: Split into BUY + optional SELL transactions  
3. **Dividends**: Convert directly to RECEIVE transactions

### Backward Compatibility
- Maintain existing import endpoints for legacy CSV formats
- Add new universal transaction import endpoint
- Provide conversion utilities for historical data

### Validation Rules
- Enforce required fields based on transaction_type
- Validate option expiration dates are future-dated for opening transactions
- Ensure quantity and price are positive for most transaction types
- Cross-validate related transactions (assignments must have corresponding options)

This format provides the flexibility needed for comprehensive testing while maintaining compatibility with Wheeler's sophisticated options trading focus.


Nominal Total: $155,313    Open Options: $4,790 / 3.1% of Nominal              Long: $55,913    Open Calls: $1,590 / 2.8% of Long                 Puts: $99,400    Open Puts: $3,200 / 3.2% of Exposure / 2.1% of Treasuries 