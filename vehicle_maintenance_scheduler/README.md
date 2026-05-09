# Vehicle Maintenance Scheduler

A production-ready scheduling and optimization service built for enterprise fleet maintenance. This service fetches available tasks and mechanic hours, computing the mathematically optimal task allocation using a 0/1 Knapsack Dynamic Programming solver.

## Architecture

This service consumes the shared `logging_middleware/pkg` SDK for all infrastructure concerns (Auth, Retries, Configuration, Observability, and Logging) and strictly isolates its own Domain Logic (`internal/scheduler`).

### Core Optimization Approach
- **Algorithm**: `0/1 Knapsack (Dynamic Programming)`
- **Objective**: Maximize `Impact Score`
- **Constraint**: `Available Mechanic Hours`
- **Why DP?**: Given the strict integer hour constraints, a dynamic programming approach guarantees the mathematically optimal selection of tasks without the overhead of heuristics. It provides a highly deterministic and resilient allocation strategy.

## Usage

This project operates within a Go workspace. 

### Local Development

1. Duplicate `.env.example` to `.env` (this is automatically shared with the middleware SDK).
2. Start the infrastructure:
   ```bash
   docker compose up -d
   ```
3. Run the Scheduler API:
   ```bash
   go run ./cmd/api
   ```

### API Endpoints
- `GET /health`: Healthcheck
- `POST /schedule`: Triggers the optimization engine and returns the optimal `OptimizationResult` payload.

## Screenshots

![Architecture Diagram](./screenshots/architecture_diagram.png)
![Scheduler Output](./screenshots/scheduler_output.png)
![Optimization Result](./screenshots/optimization_result.png)
![API Response](./screenshots/api_response.png)
