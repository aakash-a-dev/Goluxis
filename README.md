# GoLuxis

<div align="center">
  <img src="/assets/logo.png" alt="GoLuxis Logo" width="200"/>
  <h3>A Modern Go-based Redis Extension Framework</h3>
  <p>Extend Redis functionality with the power and safety of Go</p>

  [![Go Version](https://img.shields.io/github/go-mod/go-version/aakash-a-dev/Goluxis)](https://github.com/aakash-a-dev/Goluxis)
  [![License](https://img.shields.io/github/license/aakash-a-dev/Goluxis)](https://github.com/aakash-a-dev/Goluxis/blob/main/LICENSE)
  [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/aakash-a-dev/Goluxis/blob/main/CONTRIBUTING.md)
</div>

## 🚀 Features

- 🛠 Create custom Redis commands in Go
- 🏗 Define new data structures
- 🔌 Extend Redis with business-specific logic
- 🔒 Simple and safe deployment compared to C modules
- 🚦 Built-in connection management
- 📝 Full RESP protocol support

## 📦 Installation

```bash
go get github.com/aakash-a-dev/Goluxis
```

## 🎯 Quick Start

Here's a simple example of creating a custom Redis command:

```go
package main

import (
    "github.com/aakash-a-dev/Goluxis/pkg/command"
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

## 🎉 Use Cases

### 1. Custom Search Capabilities
Implement specialized search algorithms directly in Redis:
```
PRODUCTSEARCH shoes brand:nike color:red sort_by:popularity
```

### 2. Domain-Specific Data Structures
Create custom data types for specific use cases:
```
STOCKTS.ADD apple 2025-03-13 185.23
STOCKTS.MOVINGAVG apple 30d
```

### 3. Real-time Processing
Add business logic that runs directly within Redis:
```
RECOMMEND.PRODUCTS user:1234 limit:5 context:browsing
```

## 🏗 Project Status

Currently in Beta (v0.1.0-beta). Core features:
- ✅ Basic command registration and execution
- ✅ Redis protocol compatibility
- ✅ Connection management
- ✅ Error handling

Coming soon:
- 🔄 Persistence layer
- 📡 Replication support
- 🔍 Advanced data types
- 🛡 Enhanced error handling

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Redis team for their amazing work
- Go community for tools and support

## 📚 Documentation

For detailed documentation, please visit our [Wiki](https://github.com/aakash-a-dev/Goluxis/wiki).

## 🔗 Links

- [GitHub Repository](https://github.com/aakash-a-dev/Goluxis)
- [Issue Tracker](https://github.com/aakash-a-dev/Goluxis/issues)
- [Change Log](CHANGELOG.md) 