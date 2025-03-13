# Rate Limiter Example

This example demonstrates how to implement a sliding window rate limiter using GoLuxis. It provides commands for rate limiting requests and checking rate limit status.

## Features

- Sliding window rate limiting algorithm
- Configurable window size and request limits
- Real-time rate limit information
- Automatic cleanup of expired windows

## Commands

### 1. RATELIMIT.ALLOW

Check if a request is allowed under the rate limit:

```bash
RATELIMIT.ALLOW user:123 100 3600
# Arguments: key, max_requests, window_seconds
# Returns: 1 if allowed, 0 if denied
```

### 2. RATELIMIT.INFO

Get rate limit information for a key:

```bash
RATELIMIT.INFO user:123
```

## Example Usage

1. Start Redis:
```bash
docker run --name redis-test -p 6379:6379 -d redis
```

2. Build and run the example:
```bash
go build -o rate-limiter
./rate-limiter
```

3. Test rate limiting:
```bash
# Allow 5 requests per minute for user:123
redis-cli -p 6380 RATELIMIT.ALLOW "user:123" 5 60

# Check rate limit info
redis-cli -p 6380 RATELIMIT.INFO "user:123"

# Test rate limit by making multiple requests
for i in {1..6}; do
    redis-cli -p 6380 RATELIMIT.ALLOW "user:123" 5 60
done
```

## Use Cases

1. API Rate Limiting
   - Protect APIs from abuse
   - Implement fair usage policies
   - Control resource consumption

2. User Action Throttling
   - Prevent spam
   - Limit login attempts
   - Control message sending rates

3. Service Protection
   - Prevent DDoS attacks
   - Control resource usage
   - Implement tiered service levels

## Implementation Details

The rate limiter uses a sliding window algorithm:
1. Each request is recorded with its timestamp
2. When checking limits, only requests within the specified window are counted
3. Old windows are automatically cleaned up
4. Thread-safe implementation using mutexes

## Example Rate Limiting Scenarios

1. Basic API Rate Limiting:
```bash
# Allow 100 requests per hour
redis-cli -p 6380 RATELIMIT.ALLOW "api:key1" 100 3600
```

2. Login Attempt Limiting:
```bash
# Allow 5 attempts per minute
redis-cli -p 6380 RATELIMIT.ALLOW "login:user123" 5 60
```

3. Message Rate Limiting:
```bash
# Allow 10 messages per minute
redis-cli -p 6380 RATELIMIT.ALLOW "messages:user123" 10 60
``` 