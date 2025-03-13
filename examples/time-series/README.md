# Time Series Data Example

This example demonstrates how to implement time series data storage and analysis using GoLuxis. It provides commands for storing time series data points and calculating statistics.

## Features

- Store time series data points with timestamps
- Query data within a time range
- Calculate statistics (min, max, average)
- RFC3339 timestamp format support

## Commands

### 1. TS.ADD

Add a data point to a time series:

```bash
TS.ADD stock:AAPL 2025-03-14T10:00:00Z 185.23
```

### 2. TS.RANGE

Get data points within a time range:

```bash
TS.RANGE stock:AAPL 2025-03-14T00:00:00Z 2025-03-14T23:59:59Z
```

### 3. TS.STATS

Get statistics for a time series:

```bash
TS.STATS stock:AAPL
```

## Example Usage

1. Start Redis:
```bash
docker run --name redis-test -p 6379:6379 -d redis
```

2. Build and run the example:
```bash
go build -o time-series
./time-series
```

3. Add some data points:
```bash
redis-cli -p 6380 TS.ADD "stock:AAPL" "2025-03-14T10:00:00Z" "185.23"
redis-cli -p 6380 TS.ADD "stock:AAPL" "2025-03-14T10:30:00Z" "186.45"
redis-cli -p 6380 TS.ADD "stock:AAPL" "2025-03-14T11:00:00Z" "184.89"
```

4. Query data:
```bash
# Get data for a time range
redis-cli -p 6380 TS.RANGE "stock:AAPL" "2025-03-14T00:00:00Z" "2025-03-14T23:59:59Z"

# Get statistics
redis-cli -p 6380 TS.STATS "stock:AAPL"
```

## Use Cases

1. Financial Data
   - Stock prices
   - Trading volumes
   - Market indicators

2. IoT Sensors
   - Temperature readings
   - Humidity levels
   - Energy consumption

3. Application Metrics
   - Request latencies
   - Error rates
   - User activity 