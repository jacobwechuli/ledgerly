import { useState } from 'react'

type Endpoint = {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
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
        response: '{"status":"ok","service":"ledgerly-api","version":"2.0","target":"high-net-worth"}',
      },
    ],
  },
  {
    title: 'Accounts',
    icon: '🏦',
    endpoints: [
      {
        method: 'POST',
        path: '/api/v1/accounts',
        description: 'Create a new account (bank, mobile money, investment, or property)',
        auth: true,
        body: '{"name":"Chase Checking","type":"bank","sub_type":"checking","currency":"USD","balance":150000.00,"institution":"Chase Bank"}',
      },
      {
        method: 'GET',
        path: '/api/v1/accounts',
        description: 'List all accounts (multi-currency, multi-type)',
        auth: true,
      },
      {
        method: 'PUT',
        path: '/api/v1/accounts/{id}/balance',
        description: 'Update account balance',
        auth: true,
        body: '{"balance":175000.00}',
      },
      {
        method: 'DELETE',
        path: '/api/v1/accounts/{id}',
        description: 'Deactivate an account (soft delete)',
        auth: true,
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
        description: 'Record a transaction linked to a specific account',
        auth: true,
        body: '{"account_id":"uuid","amount":5000.00,"currency":"USD","type":"expense","category":"travel","description":"Business class flight to London"}',
      },
      {
        method: 'GET',
        path: '/api/v1/transactions?page=1&page_size=20',
        description: 'List transactions (paginated)',
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
    title: 'Investments',
    icon: '📈',
    endpoints: [
      {
        method: 'POST',
        path: '/api/v1/investments',
        description: 'Add an investment holding (stocks, ETFs, crypto, bonds)',
        auth: true,
        body: '{"account_id":"uuid","symbol":"AAPL","name":"Apple Inc.","type":"stock","quantity":500,"avg_cost_basis":150.00,"current_price":175.00,"currency":"USD"}',
      },
      {
        method: 'GET',
        path: '/api/v1/investments',
        description: 'List all investment holdings with unrealized gains',
        auth: true,
      },
      {
        method: 'DELETE',
        path: '/api/v1/investments/{id}',
        description: 'Remove an investment',
        auth: true,
      },
    ],
  },
  {
    title: 'Properties',
    icon: '🏠',
    endpoints: [
      {
        method: 'POST',
        path: '/api/v1/properties',
        description: 'Track real estate, vehicles, art, jewelry, or other assets',
        auth: true,
        body: '{"account_id":"uuid","name":"Nairobi Apartment","type":"real_estate","current_value":25000000.00,"purchase_price":20000000.00,"currency":"KES","location":"Westlands, Nairobi"}',
      },
      {
        method: 'GET',
        path: '/api/v1/properties',
        description: 'List all property and asset holdings',
        auth: true,
      },
      {
        method: 'DELETE',
        path: '/api/v1/properties/{id}',
        description: 'Remove a property/asset',
        auth: true,
      },
    ],
  },
  {
    title: 'Net Worth',
    icon: '💎',
    endpoints: [
      {
        method: 'GET',
        path: '/api/v1/net-worth/summary',
        description: 'Get aggregated net worth in USD with breakdown (cash, investments, property)',
        auth: true,
        response: '{"data":{"total_usd":2500000,"cash_usd":500000,"investments_usd":1500000,"property_usd":500000,"change_pct_30d":2.04}}',
      },
      {
        method: 'GET',
        path: '/api/v1/net-worth/history?days=90',
        description: 'Get daily net worth snapshots for charting',
        auth: true,
      },
    ],
  },
  {
    title: 'Currencies',
    icon: '💱',
    endpoints: [
      {
        method: 'GET',
        path: '/api/v1/currencies',
        description: 'List all supported currencies with exchange rates to USD',
        auth: true,
      },
      {
        method: 'GET',
        path: '/api/v1/currencies/convert?from=KES&to=USD&amount=100000',
        description: 'Convert between currencies',
        auth: true,
        response: '{"data":{"from_currency":"KES","to_currency":"USD","amount":100000,"result":770.00,"rate":0.0077}}',
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
        description: 'Create a recurring bill linked to an account',
        auth: true,
        body: '{"account_id":"uuid","name":"Electricity","amount":15000.00,"currency":"KES","category":"utilities","due_date":"2024-02-01T00:00:00Z","recurrence":"monthly"}',
      },
      {
        method: 'GET',
        path: '/api/v1/bills?paid=false',
        description: 'List bills (filter by paid status)',
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
    title: 'Budget Goals',
    icon: '🎯',
    endpoints: [
      {
        method: 'POST',
        path: '/api/v1/budget-goals',
        description: 'Create a large financial goal (property, education, retirement, etc.)',
        auth: true,
        body: '{"name":"Beach House in Diani","target_amount":50000000.00,"currency":"KES","category":"property","deadline":"2026-12-31T00:00:00Z","priority":"high"}',
      },
      {
        method: 'GET',
        path: '/api/v1/budget-goals',
        description: 'List all goals with progress tracking',
        auth: true,
      },
      {
        method: 'POST',
        path: '/api/v1/budget-goals/{id}/contribute',
        description: 'Add funds toward a goal',
        auth: true,
        body: '{"amount":5000000.00}',
      },
      {
        method: 'DELETE',
        path: '/api/v1/budget-goals/{id}',
        description: 'Delete a goal',
        auth: true,
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
        description: 'Create a reusable budget allocation template',
        auth: true,
        body: '{"name":"HNW Allocation","description":"High-net-worth portfolio","categories":[{"category":"investments","allocated_pct":60},{"category":"property","allocated_pct":25},{"category":"cash","allocated_pct":10},{"category":"philanthropy","allocated_pct":5}],"is_public":false}',
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
        description: 'Delete a template',
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
        path: '/api/v1/reports/expenditure?start_date=2024-01-01&end_date=2024-01-31&currency=USD',
        description: 'Spending breakdown by category (multi-currency aware)',
        auth: true,
      },
      {
        method: 'GET',
        path: '/api/v1/reports/summary',
        description: 'Total income, expenses, and net balance (all in USD)',
        auth: true,
      },
    ],
  },
  {
    title: 'Settings',
    icon: '⚙️',
    endpoints: [
      {
        method: 'GET',
        path: '/api/v1/settings',
        description: 'Get user settings (privacy mode, default currency, theme)',
        auth: true,
        response: '{"data":{"privacy_mode":true,"default_currency":"USD","theme":"dark"}}',
      },
      {
        method: 'PUT',
        path: '/api/v1/settings',
        description: 'Update user settings',
        auth: true,
        body: '{"privacy_mode":true,"default_currency":"USD","theme":"dark"}',
      },
    ],
  },
  {
    title: 'Tax',
    icon: '🧾',
    endpoints: [
      {
        method: 'GET',
        path: '/api/v1/tax/capital-gains?year=2024',
        description: 'Estimate unrealized capital gains and tax liability',
        auth: true,
        response: '{"data":{"year":2024,"unrealized_gains":250000,"estimated_tax":37500,"currency":"USD"}}',
      },
      {
        method: 'GET',
        path: '/api/v1/tax/records?year=2024',
        description: 'Get tax records for a specific year',
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
        description: 'Chat with AI about your finances (concierge tone, data-driven)',
        auth: true,
        body: '{"messages":[{"role":"user","content":"What\'s my current exposure to Kenyan Shillings?"}]}',
      },
      {
        method: 'POST',
        path: '/api/v1/ai/insights',
        description: 'Generate concierge-style financial observations (not generic tips)',
        auth: true,
        response: '{"data":[{"category":"currency","title":"USD Exposure Declined","description":"Your USD-denominated assets decreased 8% this month...","impact":"negative","priority":"high"}]}',
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
                <p className="text-xs text-gray-400">Private Wealth Management • v2.0</p>
              </div>
            </div>
            <div className="flex items-center gap-3 flex-wrap">
              <span className="px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-medium">
                Multi-Currency
              </span>
              <span className="px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-400 text-xs font-medium">
                Investments
              </span>
              <span className="px-3 py-1 rounded-full bg-purple-500/10 border border-purple-500/20 text-purple-400 text-xs font-medium">
                Net Worth
              </span>
              <span className="px-3 py-1 rounded-full bg-amber-500/10 border border-amber-500/20 text-amber-400 text-xs font-medium">
                AI Insights
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
                <h3 className="text-sm font-semibold text-gray-300 mt-4 mb-2">Currencies</h3>
                <p className="text-xs text-gray-400">USD, KES, GBP, EUR, CHF, JPY, ZAR, NGN, AED, SGD</p>
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
                                : endpoint.method === 'PUT'
                                ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
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
              <h3 className="text-lg font-bold text-white mb-4">🏗️ Architecture for High-Net-Worth Users</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="p-4 rounded-lg bg-gray-800/50 border border-gray-700">
                  <h4 className="text-sm font-semibold text-emerald-400 mb-2">Multi-Currency</h4>
                  <p className="text-xs text-gray-400">
                    First-class support for 10+ currencies. All amounts converted to USD for aggregate views. Real-time exchange rates.
                  </p>
                </div>
                <div className="p-4 rounded-lg bg-gray-800/50 border border-gray-700">
                  <h4 className="text-sm font-semibold text-blue-400 mb-2">Net Worth Tracking</h4>
                  <p className="text-xs text-gray-400">
                    Aggregate cash, investments, and property into a single net worth figure. Historical snapshots for trend analysis.
                  </p>
                </div>
                <div className="p-4 rounded-lg bg-gray-800/50 border border-gray-700">
                  <h4 className="text-sm font-semibold text-purple-400 mb-2">AI Concierge</h4>
                  <p className="text-xs text-gray-400">
                    Sophisticated insights like "USD exposure dropped 8% this month" — not generic budgeting tips or gamified nudges.
                  </p>
                </div>
              </div>
            </div>

            {/* Design Philosophy */}
            <div className="mt-6 p-6 rounded-xl bg-gradient-to-br from-gray-900 to-gray-900/50 border border-gray-800">
              <h3 className="text-lg font-bold text-white mb-4">🎯 Design Philosophy</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <h4 className="text-sm font-semibold text-gray-300 mb-2">For Sophisticated Users</h4>
                  <ul className="text-xs text-gray-400 space-y-1">
                    <li>• No gamification or motivational language</li>
                    <li>• Professional, discreet tone</li>
                    <li>• Data-driven insights with exact numbers</li>
                    <li>• Privacy mode for discretion</li>
                    <li>• Multi-account, multi-currency support</li>
                  </ul>
                </div>
                <div>
                  <h4 className="text-sm font-semibold text-gray-300 mb-2">Comprehensive Tracking</h4>
                  <ul className="text-xs text-gray-400 space-y-1">
                    <li>• Bank accounts (multiple currencies)</li>
                    <li>• Mobile money (M-Pesa, Airtel, etc.)</li>
                    <li>• Investment portfolios (stocks, crypto)</li>
                    <li>• Real estate & vehicles</li>
                    <li>• Large goals (property, education)</li>
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
            <span>Ledgerly API v2.0 • Private Wealth Management Backend</span>
            <span>Source: backend/README.md for full documentation</span>
          </div>
        </div>
      </footer>
    </div>
  )
}

export default App
