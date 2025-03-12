# GoRedis-Ext: Go-based Redis Extension Framework

GoRedis-Ext is a framework that enables developers to create Redis extensions and custom commands using Go. This project aims to simplify Redis extensibility by providing a Go-based alternative to C modules.

## Features

- Create custom Redis commands in Go
- Define new data structures
- Extend Redis with business-specific logic
- Simple and safe deployment compared to C modules

## Installation

```bash
go get github.com/goluxis/goredis-ext
```

## Quick Start

Here's a simple example of creating a custom Redis command:

```go
package main

import (
    "github.com/goluxis/goredis-ext/pkg/command"
)

func main() {
    // Create a new custom command
    cmd := command.New("HELLO.WORLD")
    
    // Define command behavior
    cmd.Handler = func(ctx *command.Context) error {
        return ctx.Reply("Hello, Redis!")
    }
    
    // Register and start the extension
    ext := command.NewExtension("hello-world")
    ext.AddCommand(cmd)
    ext.Start()
}
```

## Use Cases

1. **Custom Search Capabilities**: Implement specialized search algorithms directly in Redis
2. **Domain-Specific Data Structures**: Create custom data types for specific use cases
3. **Real-time Processing**: Add business logic that runs directly within Redis
4. **Complex Access Control**: Implement sophisticated access patterns

## Project Status

This project is currently in MVP (Minimum Viable Product) stage. Core features are being developed with a focus on:

- Basic command registration and execution
- Redis protocol compatibility
- Connection management
- Error handling

## Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Redis team for their amazing work
- Go community for tools and support 