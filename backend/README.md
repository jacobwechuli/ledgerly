# Ledgerly API Backend

A personal finance management API built with Go, PostgreSQL, and Clerk authentication.

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

### Transactions

#### Create Transaction
```
POST /api/v1/transactions
```

**Body:**
```json
{
  "amount": 150.00,
  "type": "expense",
  "category": "food",
  "description": "Grocery shopping",
  "source": "card",
  "date": "2024-01-15T10:00:00Z"
}
```

**Valid values:**
- `type`: `"income"`, `"expense"`
- `source`: `"bank"`, `"cash"`, `"mobile_money"`, `"card"`

#### List Transactions
```
GET /api/v1/transactions?page=1&page_size=20
```

#### Get Transaction
```
GET /api/v1/transactions/{id}
```

#### Delete Transaction
```
DELETE /api/v1/transactions/{id}
```

---

### Mobile Money Transactions

#### Create Mobile Money Transaction
```
POST /api/v1/mobile-money
```

**Body:**
```json
{
  "transaction_id": "QKL3ABC123",
  "amount": 500.00,
  "type": "send",
  "phone_number": "+254712345678",
  "provider": "mpesa",
  "counterparty_name": "John Doe",
  "description": "Rent payment",
  "date": "2024-01-15T10:00:00Z"
}
```

**Valid values:**
- `type`: `"send"`, `"receive"`, `"paybill"`, `"buygoods"`
- `provider`: `"mpesa"`, `"airtel_money"`, `"tigo_pesa"`

#### List Mobile Money Transactions
```
GET /api/v1/mobile-money?page=1&page_size=20
```

---

### Budget Goals

#### Create Budget Goal
```
POST /api/v1/budget-goals
```

**Body:**
```json
{
  "name": "Emergency Fund",
  "target_amount": 10000.00,
  "category": "savings",
  "deadline": "2024-12-31T00:00:00Z"
}
```

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
  "amount": 500.00
}
```

#### Delete Budget Goal
```
DELETE /api/v1/budget-goals/{id}
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
  "name": "Electricity Bill",
  "amount": 150.00,
  "category": "utilities",
  "due_date": "2024-02-01T00:00:00Z",
  "recurrence": "monthly",
  "notes": "Kenya Power"
}
```

**Valid recurrence values:** `"monthly"`, `"weekly"`, `"yearly"`, `"once"`

#### List Bills
```
GET /api/v1/bills?paid=false
```

Optional query parameter `paid` filters by payment status.

#### Mark Bill as Paid
```
POST /api/v1/bills/{id}/pay
```

#### Delete Bill
```
DELETE /api/v1/bills/{id}
```

---

### Reports

#### Expenditure Report
```
GET /api/v1/reports/expenditure?start_date=2024-01-01&end_date=2024-01-31
```

Returns spending breakdown by category for the given date range.

**Response:**
```json
{
  "data": {
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-01-31T00:00:00Z",
    "categories": [
      {"category": "food", "amount": 500.00, "count": 25},
      {"category": "transport", "amount": 200.00, "count": 15}
    ],
    "total_spent": 700.00
  }
}
```

#### Financial Summary
```
GET /api/v1/reports/summary
```

Returns total income, expenses, and net balance.

---

### Budget Templates

#### Create Budget Template
```
POST /api/v1/budget-templates
```

**Body:**
```json
{
  "name": "50/30/20 Budget",
  "description": "Classic budget allocation",
  "categories": [
    {"category": "needs", "allocated_pct": 50.00},
    {"category": "wants", "allocated_pct": 30.00},
    {"category": "savings", "allocated_pct": 20.00}
  ],
  "is_public": false
}
```

#### List Budget Templates
```
GET /api/v1/budget-templates
```

Returns user's own templates plus public templates.

#### Delete Budget Template
```
DELETE /api/v1/budget-templates/{id}
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
    {"role": "user", "content": "How much did I spend on food last month?"}
  ]
}
```

The AI has access to the user's financial context (transactions, goals, bills) and can answer questions about their finances.

#### Get Budget Suggestion
```
POST /api/v1/ai/suggest-budget
```

Analyzes the user's transaction history and generates personalized budget recommendations.

**Response:**
```json
{
  "data": {
    "categories": [
      {"category": "housing", "recommended": 1500, "percentage": 30, "rationale": "..."},
      {"category": "food", "recommended": 750, "percentage": 15, "rationale": "..."}
    ],
    "total_income": 5000,
    "notes": "Based on your spending patterns..."
  }
}
```

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

## License

Proprietary - Ledgerly
