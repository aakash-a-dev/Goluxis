# Product Search Engine Example

This example demonstrates how to create a custom product search engine using GoLuxis. It implements a simple in-memory product catalog with search and filtering capabilities.

## Features

- Add products with JSON data
- Search products by name or brand
- Filter by brand, category, and price range
- JSON response format

## Commands

### 1. PRODUCT.ADD

Add a product to the catalog:

```bash
PRODUCT.ADD product:1 '{"name": "Nike Air Max", "brand": "Nike", "category": "shoes", "price": 129.99, "tags": ["running", "sports"]}'
```

### 2. PRODUCT.SEARCH

Search products with filters:

```bash
# Basic search
PRODUCT.SEARCH nike

# Search with filters
PRODUCT.SEARCH shoes brand=nike category=running min_price=50 max_price=200
```

## Example Usage

1. Start Redis:
```bash
docker run --name redis-test -p 6379:6379 -d redis
```

2. Build and run the example:
```bash
go build -o search-engine
./search-engine
```

3. Add some products:
```bash
redis-cli -p 6380 PRODUCT.ADD "shoe1" '{"name": "Nike Air Max", "brand": "Nike", "category": "shoes", "price": 129.99, "tags": ["running", "sports"]}'
redis-cli -p 6380 PRODUCT.ADD "shoe2" '{"name": "Adidas Ultraboost", "brand": "Adidas", "category": "shoes", "price": 159.99, "tags": ["running", "sports"]}'
```

4. Search products:
```bash
redis-cli -p 6380 PRODUCT.SEARCH "nike"
redis-cli -p 6380 PRODUCT.SEARCH "running" brand=nike min_price=100
``` 