# Endpoints for Stats API

## health
> Checks health status of the API

```sh
    curl -X POST http://localhost:3000/api/v1/health
```

## info
> Checks additional info for the API

```sh
    curl -X POST http://localhost:3000/api/v1/info
```

## getStatsSummary
> Get summary stats of all runs

This endpoint will fetch the following details:

- `total_runs`: Total number of runs completed
- `total_distance`: Total distance (km) traveled for all runs
- `total_hours`: Total hours spent for the duration of all runs
- `total_climbed`: Total distance climbed (m) for all runs

```sh
    curl -X POST "http://localhost:3000/api/v1/stats/summary"
```
