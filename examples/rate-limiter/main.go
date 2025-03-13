package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/aakash-a-dev/Goluxis/pkg/command"
	"github.com/aakash-a-dev/Goluxis/pkg/resp"
)

// Window represents a time window for rate limiting
type Window struct {
	Timestamp time.Time
	Count     int64
}

// RateLimiter implements a sliding window rate limiter
type RateLimiter struct {
	windows map[string][]Window
	mu      sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		windows: make(map[string][]Window),
	}
}

func (rl *RateLimiter) cleanup(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if windows, exists := rl.windows[key]; exists {
		now := time.Now()
		var active []Window
		for _, w := range windows {
			if now.Sub(w.Timestamp) < time.Hour {
				active = append(active, w)
			}
		}
		if len(active) == 0 {
			delete(rl.windows, key)
		} else {
			rl.windows[key] = active
		}
	}
}

func main() {
	// Create rate limiter
	limiter := NewRateLimiter()

	// Create extension
	ext := command.NewExtension("rate-limiter")

	// RATELIMIT.ALLOW command
	allowCmd := command.New("RATELIMIT.ALLOW")
	allowCmd.Description = "Check if request is allowed under rate limit"
	allowCmd.Handler = func(ctx *command.Context) error {
		if len(ctx.Args) != 4 {
			return fmt.Errorf("usage: RATELIMIT.ALLOW <key> <max_requests> <window_seconds>")
		}

		key := ctx.Args[1]
		maxRequests, err := strconv.ParseInt(ctx.Args[2], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid max_requests: %v", err)
		}

		windowSeconds, err := strconv.ParseInt(ctx.Args[3], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid window_seconds: %v", err)
		}

		// Cleanup old windows
		limiter.cleanup(key)

		now := time.Now()
		windowDuration := time.Duration(windowSeconds) * time.Second

		// Calculate total requests in the current window
		limiter.mu.RLock()
		windows := limiter.windows[key]
		var totalRequests int64
		for _, w := range windows {
			if now.Sub(w.Timestamp) < windowDuration {
				totalRequests += w.Count
			}
		}
		limiter.mu.RUnlock()

		if totalRequests >= maxRequests {
			return ctx.Reply("0") // Not allowed
		}

		// Add new request to window
		limiter.mu.Lock()
		if _, exists := limiter.windows[key]; !exists {
			limiter.windows[key] = make([]Window, 0)
		}
		limiter.windows[key] = append(limiter.windows[key], Window{
			Timestamp: now,
			Count:     1,
		})
		limiter.mu.Unlock()

		return ctx.Reply("1") // Allowed
	}

	// RATELIMIT.INFO command
	infoCmd := command.New("RATELIMIT.INFO")
	infoCmd.Description = "Get rate limit information for a key"
	infoCmd.Handler = func(ctx *command.Context) error {
		if len(ctx.Args) != 2 {
			return fmt.Errorf("usage: RATELIMIT.INFO <key>")
		}

		key := ctx.Args[1]

		// Cleanup old windows
		limiter.cleanup(key)

		limiter.mu.RLock()
		windows := limiter.windows[key]
		var totalRequests int64
		now := time.Now()

		for _, w := range windows {
			if now.Sub(w.Timestamp) < time.Hour {
				totalRequests += w.Count
			}
		}
		limiter.mu.RUnlock()

		info := fmt.Sprintf(`{
			"key": "%s",
			"total_requests": %d,
			"window_count": %d
		}`, key, totalRequests, len(windows))

		return ctx.Reply(info)
	}

	// Register commands
	ext.AddCommand(allowCmd)
	ext.AddCommand(infoCmd)

	// Start TCP server
	listener, err := net.Listen("tcp", ":6380")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Printf("Rate limiter extension listening on :6380")

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
