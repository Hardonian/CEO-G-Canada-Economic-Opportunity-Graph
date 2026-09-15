// Package graphql provides a zero-dependency GraphQL-over-HTTP handler for the
// CanadaOpportunityGraph API. It implements a minimal hand-rolled executor that
// covers introspection, simple selections, and typed arguments — sufficient for
// the project/organization/event/signal/aiSovereignty/reconciliation queries
// required by the enterprise API surface.
//
// No external GraphQL libraries are used; the implementation parses the query
// string directly and routes to the appropriate resolver, keeping the module
// dependency-free beyond github.com/google/uuid.
package graphql

// Schema is the normative CEGS GraphQL SDL. It is exposed at GET /api/v1/graphql
// and is returned verbatim by the __schema introspection query.
const Schema = `
"""
CanadaOpportunityGraph Enterprise GraphQL API — CEGS 1.0
"""
schema {
  query:        Query
  subscription: Subscription
}

"""
Root query type.
"""
type Query {
  """List projects with optional filters."""
  projects(
    sector:   String
    province: String
    stage:    String
    limit:    Int
    offset:   Int
  ): [Project!]!

  """Retrieve a single project by its canonical ID or slug."""
  project(id: String!): Project

  """List organizations."""
  organizations(limit: Int): [Organization!]!

  """List recent events, optionally filtered by project."""
  events(projectId: String, limit: Int): [Event!]!

  """List momentum signals."""
  signals(limit: Int): [Signal!]!

  """Multi-jurisdiction reconciliation summary."""
  reconciliation: ReconciliationReport!

  """AI Sovereignty benchmark scores."""
  aiSovereignty: [AISovereigntyScore!]!
}

"""
Subscription webhooks for moving-project notifications.
"""
type Subscription {
  """Fires when a project transitions stage or scores change."""
  projectUpdated(id: String!): ProjectUpdate!

  """Fires when the aggregate portfolio composition changes."""
  portfolioUpdated: PortfolioUpdate!
}

# ── Domain object types ──────────────────────────────────────────────────────

type Project {
  id:             String!
  name:           String!
  slug:           String!
  sector:         String!
  province:       String!
  stage:          String!
  capexCAD:       Float!
  buildability:   Float
  investability:  Float
  updatedAt:      String!
}

type Organization {
  id:         String!
  slug:       String!
  commonName: String!
  legalName:  String
  entityType: String!
  updatedAt:  String!
}

type Event {
  id:          String!
  projectId:   String!
  eventType:   String!
  eventDate:   String!
  title:       String!
  description: String
}

type Signal {
  id:          String!
  projectId:   String!
  signalType:  String!
  strength:    Float!
  detectedAt:  String!
  description: String
}

type ReconciliationReport {
  totalRecords: Int!
  merged:       Int!
  linked:       Int!
  conflicts:    Int!
}

type AISovereigntyScore {
  entityId:    String!
  entityName:  String!
  overallScore: Float!
  tier:         String!
}

type ProjectUpdate {
  projectId:  String!
  changeType: String!
  occurredAt: String!
}

type PortfolioUpdate {
  totalProjects: Int!
  updatedAt:     String!
}
`
