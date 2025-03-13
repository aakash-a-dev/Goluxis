package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/aakash-a-dev/Goluxis/pkg/command"
	"github.com/aakash-a-dev/Goluxis/pkg/resp"
)

// TimeSeriesPoint represents a single data point
type TimeSeriesPoint struct {
	Timestamp time.Time
	Value     float64
}

// TimeSeries represents a collection of time series data
type TimeSeries struct {
	points []TimeSeriesPoint
	mu     sync.RWMutex
}

// TimeSeriesStore stores multiple time series
type TimeSeriesStore struct {
	series map[string]*TimeSeries
	mu     sync.RWMutex
}

func NewTimeSeriesStore() *TimeSeriesStore {
	return &TimeSeriesStore{
		series: make(map[string]*TimeSeries),
	}
}

func main() {
	// Create time series store
	store := NewTimeSeriesStore()

	// Create extension
	ext := command.NewExtension("time-series")

	// TS.ADD command
	addCmd := command.New("TS.ADD")
	addCmd.Description = "Add a data point to a time series"
	addCmd.Handler = func(ctx *command.Context) error {
		if len(ctx.Args) != 4 {
			return fmt.Errorf("usage: TS.ADD <key> <timestamp> <value>")
		}

		key := ctx.Args[1]
		timestamp, err := time.Parse(time.RFC3339, ctx.Args[2])
		if err != nil {
			return fmt.Errorf("invalid timestamp format, use RFC3339")
		}

		value, err := strconv.ParseFloat(ctx.Args[3], 64)
		if err != nil {
			return fmt.Errorf("invalid value: %v", err)
		}

		store.mu.Lock()
		if _, exists := store.series[key]; !exists {
			store.series[key] = &TimeSeries{
				points: make([]TimeSeriesPoint, 0),
			}
		}
		store.mu.Unlock()

		series := store.series[key]
		series.mu.Lock()
		series.points = append(series.points, TimeSeriesPoint{
			Timestamp: timestamp,
			Value:     value,
		})
		series.mu.Unlock()

		return ctx.Reply("OK")
	}

	// TS.RANGE command
	rangeCmd := command.New("TS.RANGE")
	rangeCmd.Description = "Get time series data points within a time range"
	rangeCmd.Handler = func(ctx *command.Context) error {
		if len(ctx.Args) != 4 {
			return fmt.Errorf("usage: TS.RANGE <key> <start_timestamp> <end_timestamp>")
		}

		key := ctx.Args[1]
		start, err := time.Parse(time.RFC3339, ctx.Args[2])
		if err != nil {
			return fmt.Errorf("invalid start timestamp format, use RFC3339")
		}

		end, err := time.Parse(time.RFC3339, ctx.Args[3])
		if err != nil {
			return fmt.Errorf("invalid end timestamp format, use RFC3339")
		}

		store.mu.RLock()
		series, exists := store.series[key]
		store.mu.RUnlock()

		if !exists {
			return fmt.Errorf("time series not found: %s", key)
		}

		series.mu.RLock()
		var results []string
		for _, point := range series.points {
			if point.Timestamp.After(start) && point.Timestamp.Before(end) {
				results = append(results, fmt.Sprintf("%s %.2f", point.Timestamp.Format(time.RFC3339), point.Value))
			}
		}
		series.mu.RUnlock()

		return ctx.Reply(fmt.Sprintf("[%s]", strings.Join(results, ", ")))
	}

	// TS.STATS command
	statsCmd := command.New("TS.STATS")
	statsCmd.Description = "Get statistics for a time series"
	statsCmd.Handler = func(ctx *command.Context) error {
		if len(ctx.Args) != 2 {
			return fmt.Errorf("usage: TS.STATS <key>")
		}

		key := ctx.Args[1]

		store.mu.RLock()
		series, exists := store.series[key]
		store.mu.RUnlock()

		if !exists {
			return fmt.Errorf("time series not found: %s", key)
		}

		series.mu.RLock()
		defer series.mu.RUnlock()

		if len(series.points) == 0 {
			return ctx.Reply("No data points")
		}

		// Calculate statistics
		var sum, min, max float64
		min = math.MaxFloat64
		max = -math.MaxFloat64

		for _, point := range series.points {
			sum += point.Value
			if point.Value < min {
				min = point.Value
			}
			if point.Value > max {
				max = point.Value
			}
		}

		avg := sum / float64(len(series.points))

		stats := fmt.Sprintf(`{
			"count": %d,
			"min": %.2f,
			"max": %.2f,
			"avg": %.2f
		}`, len(series.points), min, max, avg)

		return ctx.Reply(stats)
	}

	// Register commands
	ext.AddCommand(addCmd)
	ext.AddCommand(rangeCmd)
	ext.AddCommand(statsCmd)

	// Start TCP server
	listener, err := net.Listen("tcp", ":6380")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Printf("Time series extension listening on :6380")

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
