package command

import (
	"context"
	"errors"
	"sync"
)

// Common errors
var (
	ErrInvalidArgCount = errors.New("invalid number of arguments")
	ErrInvalidArgType  = errors.New("invalid argument type")
	ErrCommandNotFound = errors.New("command not found")
)

// Context represents the execution context for a Redis command
type Context struct {
	ctx     context.Context
	Args    []string
	Conn    RedisConn
	command *Command
}

// RedisConn represents a connection to Redis
type RedisConn interface {
	WriteString(s string) error
	WriteInt(i int64) error
	WriteArray(length int) error
	WriteNull() error
	WriteError(err error) error
	Flush() error
}

// HandlerFunc defines the function signature for command handlers
type HandlerFunc func(ctx *Context) error

// Command represents a Redis command
type Command struct {
	Name        string
	Handler     HandlerFunc
	MinArgs     int
	MaxArgs     int
	Description string
	mu          sync.RWMutex
}

// New creates a new Command instance
func New(name string) *Command {
	return &Command{
		Name:    name,
		MinArgs: 0,
		MaxArgs: -1, // -1 means unlimited
	}
}

// Reply sends a string response back to Redis
func (c *Context) Reply(s string) error {
	return c.Conn.WriteString(s)
}

// ReplyInt sends an integer response back to Redis
func (c *Context) ReplyInt(i int64) error {
	return c.Conn.WriteInt(i)
}

// ReplyArray starts an array response with the given length
func (c *Context) ReplyArray(length int) error {
	return c.Conn.WriteArray(length)
}

// ReplyNull sends a null response back to Redis
func (c *Context) ReplyNull() error {
	return c.Conn.WriteNull()
}

// ReplyError sends an error response back to Redis
func (c *Context) ReplyError(err error) error {
	return c.Conn.WriteError(err)
}

// Flush ensures all written data is sent to Redis
func (c *Context) Flush() error {
	return c.Conn.Flush()
}

// Extension represents a Redis extension that can contain multiple commands
type Extension struct {
	Name     string
	commands map[string]*Command
	mu       sync.RWMutex
}

// NewExtension creates a new Extension instance
func NewExtension(name string) *Extension {
	return &Extension{
		Name:     name,
		commands: make(map[string]*Command),
	}
}

// AddCommand registers a new command with the extension
func (e *Extension) AddCommand(cmd *Command) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if cmd == nil {
		return errors.New("command cannot be nil")
	}

	if cmd.Name == "" {
		return errors.New("command name cannot be empty")
	}

	if cmd.Handler == nil {
		return errors.New("command handler cannot be nil")
	}

	e.commands[cmd.Name] = cmd
	return nil
}

// GetCommand retrieves a command by name
func (e *Extension) GetCommand(name string) (*Command, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	cmd, exists := e.commands[name]
	if !exists {
		return nil, ErrCommandNotFound
	}
	return cmd, nil
}
