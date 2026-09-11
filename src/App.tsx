import { useState } from 'react'

type Endpoint = {
  method: 'GET' | 'POST' | 'DELETE'
  path: string
  description: string
  body?: string
  response?: string
  auth: boolean
}

type Section = {
  title: string
  icon: string
  endpoints: Endpoint[]
}

const sections: Section[] = [
  {
    title: 'Health',
    icon: '💚',
    endpoints: [
      {
        method: 'GET',
        path: '/health',
        description: 'Server health check. No authentication required.',
        auth: false,
        response: '{"status":"ok","service":"ledgerly-api","timestamp":"..."}',
      },
    ],
  },
  {
    title: 'Transactions',
    icon: '💳',
    endpoints: [
      {
        method: 'POST',
        path: '/api/v1/transactions',
        description: 'Create a new income or expense transaction',
        auth: true,
        body: '{"amount":150.00,"type":"expense","category":"food","description":"Groceries","source":"card","date":"2024-01-15T10:00:00Z"}',
        response: '{"data":{"id":"uuid","amount":150.00,"type":"expense",...}}',
      },
      {
        method: 'GET',
        path: '/api/v1/transactions?page=1&page_size=20',
        description: 'List all transactions (paginated)',
        auth: true,
        response: '{"data":[...],"page":1,"page_size":20,"total_count":100,"total_pages":5}',
      },
      {
        method: 'GET',
        path: '/api/v1/transactions/{id}',
        description: 'Get a specific transaction by ID',
        auth: true,
      },
      {
        method: 'DELETE',
        path: '/api/v1/transactions/{id}',
        description: 'Delete a transaction',
        auth: true,
      },
    ],
  },
  {
    title: 'Mobile Money',
    icon: '📱',
    endpoints: [
      {
        method: 'POST',
        path: '/api/v1/mobile-money',
        description: 'Record a mobile money transaction (M-Pesa, Airtel Money, etc.)',
        auth: true,
        body: '{"transaction_id":"QKL3ABC123","amount":500.00,"type":"send","phone_number":"+254712345678","provider":"mpesa","counterparty_name":"John Doe","description":"Rent payment"}',
      },
      {
        method: 'GET',
        path: '/api/v1/mobile-money?page=1&page_size=20',
        description: 'List mobile money transactions',
        auth: true,
      },
    ],
  },
  {
    title: 'Budget Goals',
    icon: '🎯',
    endpoints: [
      {
        method: 'POST',
        path: '/api/v1/budget-goals',
        description: 'Create a savings goal with target amount',
        auth: true,
        body: '{"name":"Emergency Fund","target_amount":10000.00,"category":"savings","deadline":"2024-12-31T00:00:00Z"}',
      },
      {
        method: 'GET',
        path: '/api/v1/budget-goals',
        description: 'List all budget goals with progress',
        auth: true,
      },
      {
        method: 'POST',
        path: '/api/v1/budget-goals/{id}/contribute',
        description: 'Add funds to a budget goal',
        auth: true,
        body: '{"amount":500.00}',
      },
      {
        method: 'DELETE',
        path: '/api/v1/budget-goals/{id}',
        description: 'Delete a budget goal',
        auth: true,
      },
    ],
  },
  {
    title: 'Bills',
    icon: '📋',
    endpoints: [
      {
        method: 'POST',
        path: '/api/v1/bills',
        description: 'Create a recurring bill',
        auth: true,
        body: '{"name":"Electricity","amount":150.00,"category":"utilities","due_date":"2024-02-01T00:00:00Z","recurrence":"monthly","notes":"Kenya Power"}',
      },
      {
        method: 'GET',
        path: '/api/v1/bills?paid=false',
        description: 'List bills (optionally filter by paid status)',
        auth: true,
      },
      {
        method: 'POST',
        path: '/api/v1/bills/{id}/pay',
        description: 'Mark a bill as paid',
        auth: true,
      },
      {
        method: 'DELETE',
        path: '/api/v1/bills/{id}',
        description: 'Delete a bill',
        auth: true,
      },
    ],
  },
  {
    title: 'Reports',
    icon: '📊',
    endpoints: [
      {
        method: 'GET',
        path: '/api/v1/reports/expenditure?start_date=2024-01-01&end_date=2024-01-31',
        description: 'Get spending breakdown by category for a date range',
        auth: true,
        response: '{"data":{"categories":[{"category":"food","amount":500,"count":25}],"total_spent":700}}',
      },
      {
        method: 'GET',
        path: '/api/v1/reports/summary',
        description: 'Get total income, expenses, and net balance',
        auth: true,
        response: '{"data":{"total_income":5000,"total_expense":3500,"net_balance":1500}}',
      },
    ],
  },
  {
    title: 'Budget Templates',
    icon: '📐',
    endpoints: [
      {
        method: 'POST',
        path: '/api/v1/budget-templates',
        description: 'Create a reusable budget template',
        auth: true,
        body: '{"name":"50/30/20 Budget","description":"Classic allocation","categories":[{"category":"needs","allocated_pct":50},{"category":"wants","allocated_pct":30},{"category":"savings","allocated_pct":20}],"is_public":false}',
      },
      {
        method: 'GET',
        path: '/api/v1/budget-templates',
        description: 'List templates (own + public)',
        auth: true,
      },
      {
        method: 'DELETE',
        path: '/api/v1/budget-templates/{id}',
        description: 'Delete a budget template',
        auth: true,
      },
    ],
  },
  {
    title: 'AI Assistant',
    icon: '🤖',
    endpoints: [
      {
        method: 'POST',
        path: '/api/v1/ai/chat',
        description: 'Chat with AI about your finances (has access to your transaction data)',
        auth: true,
        body: '{"messages":[{"role":"user","content":"How much did I spend on food last month?"}]}',
        response: '{"data":{"response":"You spent $500 on food last month across 25 transactions..."}}',
      },
      {
        method: 'POST',
        path: '/api/v1/ai/suggest-budget',
        description: 'Get AI-generated budget recommendations based on your spending history',
        auth: true,
        response: '{"data":{"categories":[{"category":"housing","recommended":1500,"percentage":30,"rationale":"..."}],"total_income":5000,"notes":"..."}}',
      },
    ],
  },
]

function App() {
  const [activeSection, setActiveSection] = useState<string>('Health')
  const [expandedEndpoint, setExpandedEndpoint] = useState<string | null>(null)

  const currentSection = sections.find((s) => s.title === activeSection)

  return (
    <div className="min-h-screen bg-gray-950 text-gray-100">
      {/* Header */}
      <header className="border-b border-gray-800 bg-gray-900/80 backdrop-blur-sm sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-400 to-cyan-500 flex items-center justify-center text-xl font-bold">
                L
              </div>
              <div>
                <h1 className="text-xl font-bold text-white">Ledgerly API</h1>
                <p className="text-xs text-gray-400">Personal Finance Backend • v1.0</p>
              </div>
            </div>
            <div className="flex items-center gap-4">
              <span className="px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-medium">
                Go 1.22
              </span>
              <span className="px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-medium">
                PostgreSQL
              </span>
              <span className="px-3 py-1 rounded-full bg-purple-500/10 border border-purple-500/20 text-purple-400 text-xs font-medium">
                Clerk Auth
              </span>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          {/* Sidebar Navigation */}
          <nav className="lg:col-span-1">
            <div className="sticky top-24">
              <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-4">
                Endpoints
              </h2>
              <ul className="space-y-1">
                {sections.map((section) => (
                  <li key={section.title}>
                    <button
                      onClick={() => setActiveSection(section.title)}
                      className={`w-full text-left px-4 py-2.5 rounded-lg text-sm font-medium transition-all flex items-center gap-2 ${
                        activeSection === section.title
                          ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                          : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800/50'
                      }`}
                    >
                      <span>{section.icon}</span>
                      <span>{section.title}</span>
                      <span className="ml-auto text-xs text-gray-600">
                        {section.endpoints.length}
                      </span>
                    </button>
                  </li>
                ))}
              </ul>

              {/* Quick Info */}
              <div className="mt-8 p-4 rounded-xl bg-gray-900 border border-gray-800">
                <h3 className="text-sm font-semibold text-gray-300 mb-2">Base URL</h3>
                <code className="text-xs text-emerald-400 bg-gray-800 px-2 py-1 rounded block">
                  https://api.ledgerly.com
                </code>
                <h3 className="text-sm font-semibold text-gray-300 mt-4 mb-2">Auth</h3>
                <code className="text-xs text-cyan-400 bg-gray-800 px-2 py-1 rounded block">
                  Bearer {'<clerk_jwt>'}
                </code>
                <h3 className="text-sm font-semibold text-gray-300 mt-4 mb-2">Rate Limit</h3>
                <p className="text-xs text-gray-400">60 requests/minute per user</p>
              </div>
            </div>
          </nav>

          {/* Main Content */}
          <main className="lg:col-span-3">
            {currentSection && (
              <div>
                <div className="flex items-center gap-3 mb-6">
                  <span className="text-3xl">{currentSection.icon}</span>
                  <h2 className="text-2xl font-bold text-white">{currentSection.title}</h2>
                  <span className="px-2 py-0.5 rounded bg-gray-800 text-gray-400 text-xs">
                    {currentSection.endpoints.length} endpoint{currentSection.endpoints.length > 1 ? 's' : ''}
                  </span>
                </div>

                <div className="space-y-4">
                  {currentSection.endpoints.map((endpoint, idx) => {
                    const key = `${activeSection}-${idx}`
                    const isExpanded = expandedEndpoint === key

                    return (
                      <div
                        key={key}
                        className="rounded-xl border border-gray-800 bg-gray-900/50 overflow-hidden hover:border-gray-700 transition-colors"
                      >
                        <button
                          onClick={() => setExpandedEndpoint(isExpanded ? null : key)}
                          className="w-full px-5 py-4 flex items-center gap-4 text-left"
                        >
                          <span
                            className={`px-2.5 py-1 rounded text-xs font-bold uppercase ${
                              endpoint.method === 'GET'
                                ? 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
                                : endpoint.method === 'POST'
                                ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                                : 'bg-red-500/10 text-red-400 border border-red-500/20'
                            }`}
                          >
                            {endpoint.method}
                          </span>
                          <code className="text-sm text-gray-200 font-mono flex-1">
                            {endpoint.path}
                          </code>
                          {endpoint.auth && (
                            <span className="px-2 py-0.5 rounded bg-amber-500/10 border border-amber-500/20 text-amber-400 text-xs">
                              🔒 Auth
                            </span>
                          )}
                          <svg
                            className={`w-4 h-4 text-gray-500 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                          >
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                          </svg>
                        </button>

                        {isExpanded && (
                          <div className="px-5 pb-4 border-t border-gray-800 pt-4">
                            <p className="text-sm text-gray-300 mb-4">{endpoint.description}</p>

                            {endpoint.body && (
                              <div className="mb-4">
                                <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">
                                  Request Body
                                </h4>
                                <pre className="bg-gray-950 border border-gray-800 rounded-lg p-3 text-xs text-gray-300 overflow-x-auto">
                                  <code>{JSON.stringify(JSON.parse(endpoint.body), null, 2)}</code>
                                </pre>
                              </div>
                            )}

                            {endpoint.response && (
                              <div>
                                <h4 className="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">
                                  Response
                                </h4>
                                <pre className="bg-gray-950 border border-gray-800 rounded-lg p-3 text-xs text-gray-300 overflow-x-auto">
                                  <code>{endpoint.response}</code>
                                </pre>
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    )
                  })}
                </div>
              </div>
            )}

            {/* Architecture Section */}
            <div className="mt-12 p-6 rounded-xl bg-gradient-to-br from-gray-900 to-gray-900/50 border border-gray-800">
              <h3 className="text-lg font-bold text-white mb-4">🏗️ Architecture</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="p-4 rounded-lg bg-gray-800/50 border border-gray-700">
                  <h4 className="text-sm font-semibold text-emerald-400 mb-2">Authentication</h4>
                  <p className="text-xs text-gray-400">
                    Clerk JWT verification middleware. Server-side token validation with JWKS caching. User-scoped data access enforced at the data layer.
                  </p>
                </div>
                <div className="p-4 rounded-lg bg-gray-800/50 border border-gray-700">
                  <h4 className="text-sm font-semibold text-blue-400 mb-2">Rate Limiting</h4>
                  <p className="text-xs text-gray-400">
                    Redis-backed sliding window rate limiter. Works correctly across multiple instances. 60 RPM default per user with configurable limits.
                  </p>
                </div>
                <div className="p-4 rounded-lg bg-gray-800/50 border border-gray-700">
                  <h4 className="text-sm font-semibold text-purple-400 mb-2">AI Features</h4>
                  <p className="text-xs text-gray-400">
                    OpenAI-powered financial assistant. Chat endpoint with full financial context. Budget suggestions generated from transaction history.
                  </p>
                </div>
              </div>
            </div>

            {/* Deployment Info */}
            <div className="mt-6 p-6 rounded-xl bg-gradient-to-br from-gray-900 to-gray-900/50 border border-gray-800">
              <h3 className="text-lg font-bold text-white mb-4">🚀 Deployment</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <h4 className="text-sm font-semibold text-gray-300 mb-2">Stack</h4>
                  <ul className="text-xs text-gray-400 space-y-1">
                    <li>• Go 1.22 with chi router</li>
                    <li>• PostgreSQL for data persistence</li>
                    <li>• Redis for distributed rate limiting</li>
                    <li>• Clerk for authentication</li>
                    <li>• OpenAI for AI features</li>
                  </ul>
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-gray-300 mb-2">Hosting</h4>
                  <ul className="text-xs text-gray-400 space-y-1">
                    <li>• Railway for backend deployment</li>
                    <li>• Docker-based builds</li>
                    <li>• Health check at /health</li>
                    <li>• Graceful shutdown support</li>
                    <li>• $PORT env var support</li>
                  </ul>
                </div>
              </div>
            </div>
          </main>
        </div>
      </div>

      {/* Footer */}
      <footer className="border-t border-gray-800 mt-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between text-xs text-gray-500">
            <span>Ledgerly API Backend • Built with Go</span>
            <span>Source: backend/README.md for full documentation</span>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default App
