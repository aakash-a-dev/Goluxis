# GoLuxis Examples

This directory contains practical examples demonstrating how GoLuxis extends Redis functionality. Each example shows how GoLuxis transforms basic Redis operations into powerful, specialized features.

## 1. Product Search Engine

### Before (Standard Redis):
```bash
# Store product as hash
HSET product:1 name "Nike Air Max" brand "Nike" category "shoes" price "129.99"

# Search products (limited capabilities)
KEYS "product:*"              # List all products (inefficient)
HGETALL product:1            # Get single product
```

### After (With GoLuxis):
```bash
# Add product with rich JSON data
PRODUCT.ADD product:1 '{
    "name": "Nike Air Max",
    "brand": "Nike",
    "category": "shoes",
    "price": 129.99,
    "tags": ["running", "sports"]
}'

# Advanced search with filters
PRODUCT.SEARCH "nike" brand=nike category=shoes min_price=100 max_price=200
```

## 2. Time Series Data

### Before (Standard Redis):
```bash
# Store time series data (manual implementation)
ZADD stock:AAPL 1678788000 "185.23"  # Unix timestamp
ZRANGEBYSCORE stock:AAPL 1678788000 1678874400

# No built-in statistics
```

### After (With GoLuxis):
```bash
# Add data point with human-readable timestamp
TS.ADD stock:AAPL 2025-03-14T10:00:00Z 185.23

# Query with time range
TS.RANGE stock:AAPL 2025-03-14T00:00:00Z 2025-03-14T23:59:59Z

# Get instant statistics
TS.STATS stock:AAPL
# Returns: {"count": 24, "min": 184.50, "max": 187.20, "avg": 185.85}
```

## 3. Rate Limiter

### Before (Standard Redis):
```bash
# Basic rate limiting (manual implementation)
INCR user:123:requests
EXPIRE user:123:requests 3600

# Check rate limit
GET user:123:requests
```

### After (With GoLuxis):
```bash
# Sliding window rate limiting
RATELIMIT.ALLOW user:123 100 3600  # 100 requests per hour
# Returns: 1 (allowed) or 0 (denied)

# Get detailed rate limit info
RATELIMIT.INFO user:123
# Returns: {
#   "key": "user:123",
#   "total_requests": 45,
#   "window_count": 3
# }
```

## Key Benefits

1. **Simplified Usage**
   - Complex operations wrapped in simple commands
   - Human-readable parameters and responses
   - JSON support for structured data

2. **Enhanced Features**
   - Advanced filtering and search capabilities
   - Built-in statistical analysis
   - Sliding window rate limiting

3. **Better Developer Experience**
   - Clear command structure
   - Rich response formats
   - Automatic data management

## Getting Started

1. Start Redis:
```bash
docker run --name redis-test -p 6379:6379 -d redis
```

2. Choose an example and build it:
```bash
cd examples/[example-name]
go build
./[example-name]
```

3. Try the commands using redis-cli:
```bash
redis-cli -p 6380 [COMMAND] [ARGS...]
```

## Directory Structure

```
examples/
├── search-engine/     # Product search implementation
├── time-series/      # Time series data handling
└── rate-limiter/     # Rate limiting service
```

Each example directory contains:
- `main.go` - Implementation
- `README.md` - Detailed documentation
- Example commands and use cases 