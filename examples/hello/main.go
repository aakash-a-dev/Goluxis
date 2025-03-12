package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/goluxis/goredis-ext/pkg/command"
	"github.com/goluxis/goredis-ext/pkg/resp"
)

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
	return nil // Writer already flushes after each write
}

func main() {
	// Create a new extension
	ext := command.NewExtension("hello-world")

	// Create the HELLO.WORLD command
	helloCmd := command.New("HELLO.WORLD")
	helloCmd.Description = "Returns a greeting message"
	helloCmd.Handler = func(ctx *command.Context) error {
		if len(ctx.Args) > 1 {
			return ctx.Reply(fmt.Sprintf("Hello, %s!", ctx.Args[1]))
		}
		return ctx.Reply("Hello, World!")
	}

	// Register the command
	if err := ext.AddCommand(helloCmd); err != nil {
		log.Fatalf("Failed to register command: %v", err)
	}

	// Start TCP server
	listener, err := net.Listen("tcp", ":6380")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Printf("Redis extension server listening on :6380")

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
