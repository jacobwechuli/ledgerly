# Ledgerly API Backend

A sophisticated personal wealth management API built with Go, PostgreSQL, and Clerk authentication. Designed for high-net-worth users with multi-currency support, investment tracking, property management, and AI-powered financial insights.

## Architecture

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── auth/            # Clerk JWT verification middleware
│   ├── config/          # Environment configuration
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # Rate limiting (Redis-backed)
│   ├── models/          # Data models and request/response types
│   └── store/           # PostgreSQL data access layer
├── migrations/          # Database schema migrations
├── Dockerfile           # Multi-stage Docker build
└── railway.json         # Railway deployment config
```

## Key Features

### For High-Net-Worth Users
- **Multi-Account Aggregation** — Track multiple bank accounts, mobile money wallets, investment portfolios, and property assets
- **Multi-Currency Support** — First-class support for USD, KES, GBP, EUR, CHF, JPY, ZAR, NGN, AED, SGD with real-time conversion
- **Net Worth Tracking** — Aggregate cash, investments, and property into a single USD-denominated net worth figure, tracked over time
- **Investment Portfolio** — Track stocks, ETFs, mutual funds, crypto, and bonds with unrealized gains/losses
- **Property & Asset Tracking** — Real estate, vehicles, art, jewelry with purchase price and current valuation
- **Privacy/Discreet Mode** — User-level setting to mask balances by default (for iPhone users who value discretion)
- **Tax Liability Estimates** — Capital gains calculations and tax record tracking
- **AI-Powered Insights** — Concierge-style observations (not generic budgeting tips) like "USD exposure dropped 8% this month"

### Technical
- **Clerk Authentication** — Server-side JWT verification on every protected route
- **Redis Rate Limiting** — Distributed sliding window rate limiter that works across multiple instances
- **PostgreSQL** — Robust relational database with proper indexing
- **Graceful Shutdown** — Proper signal handling for production deployments
- **Health Checks** — `/health` endpoint for monitoring

## Prerequisites

- Go 1.22+
- PostgreSQL 14+
- Redis 7+ (for rate limiting across instances)
- Clerk account (for authentication)
- OpenAI API key (for AI features)

## Setup

### 1. Clone and install dependencies

```bash
cd backend
go mod download
```

### 2. Configure environment variables

Copy `.env.example` to `.env` and fill in the values:

```bash
cp .env.example .env
```

### 3. Set up the database

Run the migration SQL against your PostgreSQL instance:

```bash
psql $DATABASE_URL -f migrations/001_initial_schema.sql
```

### 4. Run the server

```bash
go run ./cmd/server/main.go
```

The server will start on `http://localhost:8080` (or the port specified in `$PORT`).

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `PORT` | No | Server port (default: `8080`) |
| `DATABASE_URL` | **Yes** | PostgreSQL connection string |
| `REDIS_URL` | No | Redis URL for distributed rate limiting |
| `CLERK_SECRET_KEY` | Prod | Clerk secret key for JWT verification |
| `CLERK_ISSUER_URL` | No | Clerk issuer URL for JWT validation |
| `OPENAI_API_KEY` | No | OpenAI API key for AI assistant features |
| `ENVIRONMENT` | No | `development` or `production` (default: `development`) |
| `RATE_LIMIT_RPM` | No | Requests per minute limit (default: `60`) |

## API Endpoints

### Health Check

```
GET /health
```

Returns server status. No authentication required.

**Response:**
```json
{
  "status": "ok",
  "service": "ledgerly-api",
  "version": "2.0",
  "target": "high-net-worth",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

---

### Authentication

All `/api/v1/*` endpoints require a valid Clerk JWT in the `Authorization` header:

```
Authorization: Bearer <clerk_jwt_token>
```

---

### Accounts (Multi-Account Support)

#### Create Account
```
POST /api/v1/accounts
```

**Body:**
```json
{
  "name": "Chase Checking",
  "type": "bank",
  "sub_type": "checking",
  "currency": "USD",
  "balance": 150000.00,
  "institution": "Chase Bank"
}
```

**Valid values:**
- `type`: `"bank"`, `"mobile_money"`, `"investment"`, `"property"`
- `sub_type`: `"checking"`, `"savings"`, `"mpesa"`, `"stock_portfolio"`, `"crypto_wallet"`, `"real_estate"`, `"vehicle"`

#### List Accounts
```
GET /api/v1/accounts
```

#### Update Account Balance
```
PUT /api/v1/accounts/{id}/balance
```

**Body:**
```json
{
  "balance": 175000.00
}
```

#### Delete Account (soft delete)
```
DELETE /api/v1/accounts/{id}
```

---

### Transactions

#### Create Transaction
```
POST /api/v1/transactions
```

**Body:**
```json
{
  "account_id": "uuid",
  "amount": 5000.00,
  "currency": "USD",
  "type": "expense",
  "category": "travel",
  "description": "Business class flight to London",
  "date": "2024-01-15T10:00:00Z"
}
```

**Valid values:**
- `type`: `"income"`, `"expense"`, `"transfer"`

#### List Transactions
```
GET /api/v1/transactions?page=1&page_size=20
```

#### Delete Transaction
```
DELETE /api/v1/transactions/{id}
```

---

### Investments (Stocks, Funds, Crypto)

#### Create Investment
```
POST /api/v1/investments
```

**Body:**
```json
{
  "account_id": "uuid",
  "symbol": "AAPL",
  "name": "Apple Inc.",
  "type": "stock",
  "quantity": 500,
  "avg_cost_basis": 150.00,
  "current_price": 175.00,
  "currency": "USD"
}
```

**Valid values:**
- `type`: `"stock"`, `"etf"`, `"mutual_fund"`, `"crypto"`, `"bond"`

#### List Investments
```
GET /api/v1/investments
```

#### Delete Investment
```
DELETE /api/v1/investments/{id}
```

---

### Properties & Assets

#### Create Property
```
POST /api/v1/properties
```

**Body:**
```json
{
  "account_id": "uuid",
  "name": "Nairobi Apartment",
  "type": "real_estate",
  "description": "3-bedroom apartment in Westlands",
  "current_value": 25000000.00,
  "purchase_price": 20000000.00,
  "currency": "KES",
  "location": "Westlands, Nairobi",
  "purchase_date": "2020-06-15T00:00:00Z"
}
```

**Valid values:**
- `type`: `"real_estate"`, `"vehicle"`, `"art"`, `"jewelry"`, `"other"`

#### List Properties
```
GET /api/v1/properties
```

#### Delete Property
```
DELETE /api/v1/properties/{id}
```

---

### Net Worth

#### Get Net Worth Summary
```
GET /api/v1/net-worth/summary
```

Returns aggregated net worth in USD with breakdown by category.

**Response:**
```json
{
  "data": {
    "total_usd": 2500000.00,
    "cash_usd": 500000.00,
    "investments_usd": 1500000.00,
    "property_usd": 500000.00,
    "change_30d": 50000.00,
    "change_pct_30d": 2.04
  }
}
```

#### Get Net Worth History
```
GET /api/v1/net-worth/history?days=90
```

Returns daily net worth snapshots for charting.

---

### Currencies

#### List Currencies
```
GET /api/v1/currencies
```

Returns all supported currencies with exchange rates to USD.

#### Convert Currency
```
GET /api/v1/currencies/convert?from=KES&to=USD&amount=100000
```

**Response:**
```json
{
  "data": {
    "from_currency": "KES",
    "to_currency": "USD",
    "amount": 100000.00,
    "result": 770.00,
    "rate": 0.0077
  }
}
```

---

### Bills

#### Create Bill
```
POST /api/v1/bills
```

**Body:**
```json
{
  "account_id": "uuid",
  "name": "Electricity Bill",
  "amount": 15000.00,
  "currency": "KES",
  "category": "utilities",
  "due_date": "2024-02-01T00:00:00Z",
  "recurrence": "monthly",
  "notes": "Kenya Power"
}
```

#### List Bills
```
GET /api/v1/bills?paid=false
```

#### Mark Bill as Paid
```
POST /api/v1/bills/{id}/pay
```

#### Delete Bill
```
DELETE /api/v1/bills/{id}
```

---

### Budget Goals (Large Goals)

#### Create Budget Goal
```
POST /api/v1/budget-goals
```

**Body:**
```json
{
  "name": "Beach House in Diani",
  "target_amount": 50000000.00,
  "currency": "KES",
  "category": "property",
  "deadline": "2026-12-31T00:00:00Z",
  "priority": "high"
}
```

**Valid values:**
- `category`: `"property"`, `"education"`, `"retirement"`, `"travel"`, `"business"`
- `priority`: `"high"`, `"medium"`, `"low"`

#### List Budget Goals
```
GET /api/v1/budget-goals
```

#### Contribute to Goal
```
POST /api/v1/budget-goals/{id}/contribute
```

**Body:**
```json
{
  "amount": 5000000.00
}
```

#### Delete Budget Goal
```
DELETE /api/v1/budget-goals/{id}
```

---

### Budget Templates

#### Create Budget Template
```
POST /api/v1/budget-templates
```

**Body:**
```json
{
  "name": "HNW Allocation",
  "description": "High-net-worth portfolio allocation",
  "categories": [
    {"category": "investments", "allocated_pct": 60.00},
    {"category": "property", "allocated_pct": 25.00},
    {"category": "cash", "allocated_pct": 10.00},
    {"category": "philanthropy", "allocated_pct": 5.00}
  ],
  "is_public": false
}
```

#### List Budget Templates
```
GET /api/v1/budget-templates
```

#### Delete Budget Template
```
DELETE /api/v1/budget-templates/{id}
```

---

### Reports

#### Expenditure Report
```
GET /api/v1/reports/expenditure?start_date=2024-01-01&end_date=2024-01-31&currency=USD
```

Returns spending breakdown by category, optionally converted to a specific currency.

**Response:**
```json
{
  "data": {
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-01-31T00:00:00Z",
    "categories": [
      {"category": "travel", "amount": 15000.00, "count": 3},
      {"category": "dining", "amount": 5000.00, "count": 12}
    ],
    "total_spent": 20000.00,
    "currency": "USD"
  }
}
```

#### Financial Summary
```
GET /api/v1/reports/summary
```

Returns total income, expenses, and net balance (all in USD).

---

### User Settings

#### Get Settings
```
GET /api/v1/settings
```

Returns user preferences including privacy mode.

**Response:**
```json
{
  "data": {
    "user_id": "user_123",
    "privacy_mode": true,
    "default_currency": "USD",
    "theme": "dark",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

#### Update Settings
```
PUT /api/v1/settings
```

**Body:**
```json
{
  "privacy_mode": true,
  "default_currency": "USD",
  "theme": "dark"
}
```

---

### Tax

#### Estimate Capital Gains
```
GET /api/v1/tax/capital-gains?year=2024
```

Returns unrealized capital gains and estimated tax liability.

**Response:**
```json
{
  "data": {
    "year": 2024,
    "unrealized_gains": 250000.00,
    "estimated_tax_rate": 0.15,
    "estimated_tax": 37500.00,
    "currency": "USD",
    "note": "Estimate based on 15% long-term capital gains rate. Consult a tax professional for accurate figures."
  }
}
```

#### Get Tax Records
```
GET /api/v1/tax/records?year=2024
```

---

### AI Assistant

#### Chat with AI
```
POST /api/v1/ai/chat
```

**Body:**
```json
{
  "messages": [
    {"role": "user", "content": "What's my current exposure to Kenyan Shillings?"}
  ]
}
```

The AI has access to the user's complete financial context (accounts, transactions, investments, properties) and provides sophisticated, data-driven answers in a professional tone.

#### Get AI Insights
```
POST /api/v1/ai/insights
```

Generates concierge-style financial observations (not generic budgeting tips).

**Response:**
```json
{
  "data": [
    {
      "id": "insight_1",
      "category": "currency",
      "title": "USD Exposure Declined",
      "description": "Your USD-denominated assets decreased by 8% this month due to KES strengthening against the dollar. Consider rebalancing if USD exposure is a strategic priority.",
      "impact": "negative",
      "priority": "high",
      "generated_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": "insight_2",
      "category": "portfolio",
      "title": "Tech Concentration Risk",
      "description": "Technology stocks represent 45% of your equity portfolio, exceeding typical diversification guidelines. Consider reducing exposure to AAPL and MSFT.",
      "impact": "negative",
      "priority": "high",
      "generated_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

**Insight categories:** `portfolio`, `cash_flow`, `currency`, `tax`, `goals`

---

## Rate Limiting

All API routes are rate-limited using a Redis-backed sliding window algorithm. Default limit is 60 requests per minute per user.

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Remaining requests in window
- `X-RateLimit-Reset`: Unix timestamp when the window resets

When the limit is exceeded, the API returns `429 Too Many Requests`.

## Deployment

### Railway

The project is configured for Railway deployment:

1. Connect your GitHub repository
2. Set the root directory to `backend/`
3. Add environment variables in Railway dashboard
4. Add PostgreSQL and Redis services
5. Deploy!

The `railway.json` configures:
- Docker-based builds
- Health check at `/health`
- Automatic restarts on failure

### Manual Deployment

```bash
# Build Docker image
docker build -t ledgerly-api .

# Run container
docker run -p 8080:8080 \
  -e DATABASE_URL="postgresql://..." \
  -e REDIS_URL="redis://..." \
  -e CLERK_SECRET_KEY="sk_..." \
  -e OPENAI_API_KEY="sk-..." \
  ledgerly-api
```

## Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test -v ./internal/auth/...
go test -v ./internal/store/...
```

## Security

- **Authentication**: Clerk JWT verification on all protected routes
- **Authorization**: User-scoped data access (server enforces user_id filtering)
- **Rate Limiting**: Redis-backed distributed rate limiting
- **Input Validation**: Request body validation using go-playground/validator
- **CORS**: Configurable CORS middleware
- **SQL Injection**: Parameterized queries via pgx
- **Privacy Mode**: User-level setting to mask sensitive balances

## Design Philosophy

Ledgerly is built for sophisticated users who value:
- **Discretion** — No gamification, no emojis, no motivational language
- **Precision** — Exact numbers, specific percentages, data-driven insights
- **Multi-currency** — Seamless handling of USD, KES, GBP, EUR, and more
- **Comprehensive tracking** — Cash, investments, property, all in one place
- **Professional tone** — Like a senior relationship manager at a private bank

## License

Proprietary - Ledgerly
