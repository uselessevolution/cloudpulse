# CloudPulse

CloudPulse is a cloud service monitoring and AI-assisted incident analysis platform built with Go, PostgreSQL, Prometheus, React, and TypeScript.

It continuously checks configured service endpoints, records health-check results, tracks service runtime state, automatically creates and resolves incidents, exposes Prometheus metrics, and provides AI-assisted incident summaries and remediation suggestions.

## Highlights

- Concurrent HTTP health checks using a bounded Go worker pool
- Database-backed per-service scheduling
- Automatic `HEALTHY → DEGRADED → DOWN → HEALTHY` state transitions
- Automatic incident creation, update, and resolution
- Prometheus metrics and structured logs
- AI incident analysis through a provider-agnostic interface
- React + TypeScript monitoring dashboard

## Why I Built This

CloudPulse was designed as a portfolio project focused on backend engineering, site reliability engineering, observability, concurrency, and incident management.

The project intentionally separates deterministic monitoring logic from AI-assisted analysis:

- CloudPulse backend decides whether a service is `HEALTHY`, `DEGRADED`, or `DOWN`.
- AI does not control health state.
- AI receives structured incident context and provides summaries, possible causes, and recommended actions.

This keeps monitoring behavior predictable while still demonstrating practical AI integration.

## Tech Stack

### Backend

- Go
- Chi Router
- PostgreSQL
- pgx
- Docker Compose

### Monitoring and Observability

- Concurrent worker pool
- Per-service monitoring schedule
- Prometheus
- Structured application logs

### AI

- Provider-agnostic `IncidentAnalyzer` interface
- Mock analyzer for free local development and demonstrations
- OpenAI analyzer for real model-backed incident analysis

### Frontend

- React
- TypeScript
- Vite
- Axios
- React Router

## Core Features

### Service Management

CloudPulse provides REST APIs for managing monitored services.

Each service defines:

- Name
- URL
- Expected HTTP status
- Monitoring interval
- Request timeout
- Enabled state

### Concurrent Health Monitoring

A background scheduler identifies services that are due for checks.

Health checks are executed through a bounded worker pool to avoid unbounded goroutine creation.

Each result records:

- HTTP status
- Latency
- Success or failure
- Error information
- Check timestamp

### Service Runtime State

CloudPulse maintains three runtime states:

```text
HEALTHY
DEGRADED
DOWN
```

Current transition rules:

```text
First failure
→ DEGRADED

Three consecutive failures
→ DOWN

Two consecutive successful checks after failure
→ HEALTHY
```

### Incident Lifecycle

When a service reaches the `DOWN` state, CloudPulse automatically creates an incident.

Only one `OPEN` incident is allowed per service.

Further failures update the existing incident rather than creating duplicates.

After sufficient successful health checks, CloudPulse automatically resolves the incident.

### Prometheus Metrics

CloudPulse exposes:

```text
GET /metrics
```

Custom metrics include:

```text
cloudpulse_health_checks_total
cloudpulse_health_check_latency_seconds
cloudpulse_open_incidents
```

### AI-Assisted Incident Analysis

CloudPulse can generate an analysis for an incident through:

```text
POST /api/incidents/{id}/ai-summary
```

The analysis contains:

```json
{
  "summary": "...",
  "possibleCauses": [
    "..."
  ],
  "recommendedActions": [
    "..."
  ]
}
```

The AI receives structured context including:

- Incident information
- Service information
- Runtime status
- Failure count
- Last recorded error
- Recent health-check history

AI analysis is intentionally separated from deterministic service-state decisions.

## AI Provider Modes

CloudPulse supports multiple incident-analysis implementations through the `IncidentAnalyzer` interface.

### Mock Mode

Default:

```text
AI_PROVIDER=mock
```

This mode:

- Requires no external API
- Requires no API key
- Costs nothing
- Is suitable for local development and portfolio demonstrations

### OpenAI Mode

```text
AI_PROVIDER=openai
OPENAI_MODEL=<model>
OPENAI_API_KEY=<your-api-key>
```

The OpenAI API key should never be committed to Git.

## Architecture
![CloudPulse Architecture](docs/architecture.png)

## Screenshots

### Service Monitoring Dashboard

![CloudPulse Services Dashboard](docs/screenshots/services-dashboard.png)

### Incident Management

![CloudPulse Incidents Dashboard](docs/screenshots/incidents-dashboard.png)

### AI-Assisted Incident Analysis

![CloudPulse AI Incident Analysis](docs/screenshots/ai-incident-analysis.png)

## Project Structure

```text
cloudpulse/
├── cmd/
│   └── api/
├── internal/
│   ├── ai/
│   ├── config/
│   ├── database/
│   ├── healthcheck/
│   ├── httpapi/
│   ├── incident/
│   ├── metrics/
│   ├── monitoring/
│   └── service/
├── migrations/
├── frontend/
├── docker-compose.yml
├── prometheus.yml
├── go.mod
└── README.md
```

## API Overview

### Health

```text
GET /health
```

### Services

```text
POST   /api/services/
GET    /api/services/
GET    /api/services/{id}
DELETE /api/services/{id}
```

### Incidents

```text
GET  /api/incidents/
GET  /api/incidents/{id}
POST /api/incidents/{id}/ai-summary
```

### Metrics

```text
GET /metrics
```

## Running Locally

### Requirements

Install:

- Go
- Node.js
- Docker Desktop

### 1. Start PostgreSQL and Prometheus

From the project root:

```powershell
docker compose up -d
```

Check:

```powershell
docker compose ps
```

### 2. Start the Go Backend

PowerShell:

```powershell
$env:DATABASE_URL="postgres://cloudpulse:cloudpulse_dev@localhost:5432/cloudpulse"
$env:AI_PROVIDER="mock"
go run .\cmd\api
```

Backend:

```text
http://localhost:8080
```

### 3. Start the React Frontend

Open another PowerShell:

```powershell
cd frontend
npm install
npm run dev
```

Frontend:

```text
http://localhost:5173
```

### 4. Prometheus

```text
http://localhost:9090
```

## Testing

Backend:

```powershell
go test ./...
go vet ./...
```

Frontend:

```powershell
cd frontend
npm run lint
npm run build
```

## Engineering Highlights

CloudPulse demonstrates several backend and SRE-oriented engineering concepts:

- Concurrent background job execution using a bounded worker pool
- Database-backed scheduling
- Health-check persistence and latency measurement
- Explicit runtime-state transitions
- Automated incident creation and recovery
- Database constraint preventing duplicate open incidents
- Prometheus instrumentation
- Structured operational logging
- Graceful application shutdown
- Dependency inversion through Go interfaces
- Provider abstraction for AI integrations
- React frontend consuming a Go REST API
- CORS configuration between frontend and backend

## Future Improvements

Potential extensions include:

- Authentication and multi-user support
- Notification channels such as email or Slack
- Historical latency charts
- Configurable incident thresholds
- Retry and backoff policies
- Distributed monitoring agents
- Alert routing and escalation policies
- Containerized backend/frontend deployment
- CI/CD pipeline
