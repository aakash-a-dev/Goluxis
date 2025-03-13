package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/aakash-a-dev/Goluxis/pkg/command"
	"github.com/aakash-a-dev/Goluxis/pkg/resp"
)

// Product represents a product in our catalog
type Product struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Brand    string   `json:"brand"`
	Category string   `json:"category"`
	Price    float64  `json:"price"`
	Tags     []string `json:"tags"`
	Score    float64  `json:"score"`
}

// ProductStore is our in-memory product database
type ProductStore struct {
	products map[string]Product
	mu       sync.RWMutex
}

func NewProductStore() *ProductStore {
	return &ProductStore{
		products: make(map[string]Product),
	}
}

func main() {
	// Create product store
	store := NewProductStore()

	// Create extension
	ext := command.NewExtension("product-search")

	// PRODUCT.ADD command
	addCmd := command.New("PRODUCT.ADD")
	addCmd.Description = "Add a product to the catalog"
	addCmd.Handler = func(ctx *command.Context) error {
		if len(ctx.Args) != 3 {
			return fmt.Errorf("usage: PRODUCT.ADD <id> <json_data>")
		}

		id := ctx.Args[1]
		jsonData := ctx.Args[2]

		var product Product
		if err := json.Unmarshal([]byte(jsonData), &product); err != nil {
			return fmt.Errorf("invalid JSON: %v", err)
		}

		product.ID = id
		store.mu.Lock()
		store.products[id] = product
		store.mu.Unlock()

		return ctx.Reply("OK")
	}

	// PRODUCT.SEARCH command
	searchCmd := command.New("PRODUCT.SEARCH")
	searchCmd.Description = "Search products with filters"
	searchCmd.Handler = func(ctx *command.Context) error {
		if len(ctx.Args) < 2 {
			return fmt.Errorf("usage: PRODUCT.SEARCH <query> [brand=X] [category=Y] [min_price=N] [max_price=M]")
		}

		query := strings.ToLower(ctx.Args[1])
		filters := make(map[string]string)

		// Parse filters
		for _, arg := range ctx.Args[2:] {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				filters[strings.ToLower(parts[0])] = strings.ToLower(parts[1])
			}
		}

		// Search and filter products
		var results []Product
		store.mu.RLock()
		for _, product := range store.products {
			// Basic search matching
			if !strings.Contains(strings.ToLower(product.Name), query) &&
				!strings.Contains(strings.ToLower(product.Brand), query) {
				continue
			}

			// Apply filters
			if brand, ok := filters["brand"]; ok && strings.ToLower(product.Brand) != brand {
				continue
			}
			if category, ok := filters["category"]; ok && strings.ToLower(product.Category) != category {
				continue
			}
			if minPrice, ok := filters["min_price"]; ok {
				if price, err := strconv.ParseFloat(minPrice, 64); err == nil && product.Price < price {
					continue
				}
			}
			if maxPrice, ok := filters["max_price"]; ok {
				if price, err := strconv.ParseFloat(maxPrice, 64); err == nil && product.Price > price {
					continue
				}
			}

			results = append(results, product)
		}
		store.mu.RUnlock()

		// Convert results to JSON
		jsonResults, err := json.Marshal(results)
		if err != nil {
			return err
		}

		return ctx.Reply(string(jsonResults))
	}

	// Register commands
	ext.AddCommand(addCmd)
	ext.AddCommand(searchCmd)

	// Start TCP server
	listener, err := net.Listen("tcp", ":6380")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Printf("Product search engine listening on :6380")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		listener.Close()
	}()

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
				break
			}
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(conn, ext)
	}
}

func handleConnection(conn net.Conn, ext *command.Extension) {
	defer conn.Close()

	reader := resp.NewReader(conn)
	writer := resp.NewWriter(conn)
	rConn := &redisConn{writer: writer}

	for {
		// Read command
		obj, err := reader.ReadObject()
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading command: %v", err)
			}
			return
		}

		// Parse command array
		cmdArray, ok := obj.([]interface{})
		if !ok {
			rConn.WriteError(fmt.Errorf("invalid command format"))
			continue
		}

		if len(cmdArray) == 0 {
			rConn.WriteError(fmt.Errorf("empty command"))
			continue
		}

		// Get command name
		cmdName, ok := cmdArray[0].(string)
		if !ok {
			rConn.WriteError(fmt.Errorf("invalid command name"))
			continue
		}

		// Get command
		cmd, err := ext.GetCommand(cmdName)
		if err != nil {
			rConn.WriteError(err)
			continue
		}

		// Convert arguments to strings
		args := make([]string, len(cmdArray))
		for i, arg := range cmdArray {
			args[i] = fmt.Sprint(arg)
		}

		// Create context
		ctx := &command.Context{
			Args: args,
			Conn: rConn,
		}

		// Execute command
		if err := cmd.Handler(ctx); err != nil {
			rConn.WriteError(err)
		}
	}
}

type redisConn struct {
	writer *resp.Writer
}

func (c *redisConn) WriteString(s string) error {
	return c.writer.WriteBulkString(s)
}

func (c *redisConn) WriteInt(i int64) error {
	return c.writer.WriteInteger(i)
}

func (c *redisConn) WriteArray(length int) error {
	return c.writer.WriteArray(length)
}

func (c *redisConn) WriteNull() error {
	return c.writer.WriteBulkString("")
}

func (c *redisConn) WriteError(err error) error {
	return c.writer.WriteError(err)
}

func (c *redisConn) Flush() error {
	return nil
}
