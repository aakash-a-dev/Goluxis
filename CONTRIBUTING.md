# Contributing to GoRedis-Ext

Thank you for your interest in contributing to GoRedis-Ext! This document provides guidelines and instructions for beta testers and contributors.

## Beta Testing

### Prerequisites

- Go 1.21 or later
- Redis CLI (for testing)
- Make (optional, for using the Makefile)

### Getting Started

1. Clone the repository:
```bash
git clone https://github.com/goluxis/goredis-ext
cd goredis-ext
```

2. Build the project:
```bash
make build
```

3. Run tests:
```bash
make test
```

4. Try the example:
```bash
make run-example
```

5. In another terminal, test with redis-cli:
```bash
redis-cli -p 6380
> HELLO.WORLD
"Hello, World!"
```

### Reporting Issues

When reporting issues, please include:

1. Go version (`go version`)
2. Operating system and version
3. Steps to reproduce the issue
4. Expected vs actual behavior
5. Any relevant error messages or logs

### Beta Testing Focus Areas

Please pay special attention to:

1. Command Registration
   - Creating custom commands
   - Command argument handling
   - Error handling

2. Redis Protocol
   - RESP protocol compatibility
   - Data type handling
   - Connection stability

3. Performance
   - Memory usage
   - CPU usage
   - Connection handling

## Contributing Code

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting:
```bash
make test
make lint
```
5. Submit a pull request

### Code Style

- Follow standard Go conventions
- Use gofmt for formatting
- Add tests for new features
- Update documentation as needed

### Pull Request Process

1. Ensure all tests pass
2. Update relevant documentation
3. Add your changes to CHANGELOG.md
4. Request review from maintainers

## License

By contributing, you agree that your contributions will be licensed under the project's MIT License. 