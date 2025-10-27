# Go Utils Library

A comprehensive Go utilities library providing common functionality for slice operations, random number generation, HTTP client, logging, and WebSocket communication.

## Features

- **Slice Operations**: Chainable slice operations with generic support
- **Random Generation**: Flexible random number and string generation
- **HTTP Client**: Easy-to-use HTTP client with chainable configuration
- **Logging**: Structured logging with file rotation and colored output
- **WebSocket**: Complete WebSocket client and server hub implementation

## Installation

```bash
go get github.com/so68/utils
```

## Modules

### 1. Slice Operations (`slice.go`)

Provides chainable slice operations with generic type support.

```go
package main

import "github.com/so68/utils"

func main() {
    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    
    result := utils.NewSlice(numbers).
        Filter(func(x int) bool { return x%2 == 0 }).
        Map(func(x int) int { return x * 2 }).
        ToSlice()
    
    fmt.Println(result) // [4, 8, 12, 16, 20]
}
```

**Key Features:**
- Generic type support (`Slice[T]`)
- Chainable operations
- Filter, Map, Reduce, Sort, Unique operations
- Batch operations and pagination
- Type conversion utilities

### 2. Random Generation (`random.go`)

Comprehensive random number and string generation utilities.

```go
package main

import "github.com/so68/utils"

func main() {
    rg := utils.NewRandomGenerator()
    
    // Generate random integers
    num := rg.IntRange(1, 100)
    
    // Generate random strings
    str := rg.String(10)
    
    // Generate random passwords
    password := rg.Password(12, true, true, true)
}
```

**Key Features:**
- Integer and float64 random generation
- String and password generation
- UUID generation
- Weighted random selection
- Seed-based generation for reproducible results

### 3. HTTP Client (`http.go`)

Easy-to-use HTTP client with chainable configuration.

```go
package main

import "github.com/so68/utils"

func main() {
    client := utils.NewHTTPClient("https://api.example.com").
        SetTimeout(10 * time.Second).
        SetHeader("Authorization", "Bearer token")
    
    response := client.Get("/users")
    if response.Error != nil {
        log.Fatal(response.Error)
    }
    
    fmt.Println(string(response.Body))
}
```

**Key Features:**
- Chainable configuration
- JSON request/response handling
- File upload support
- Custom headers and timeouts
- Error handling

### 4. Logging (`logger/`)

Structured logging with file rotation and colored console output.

```go
package main

import (
    "github.com/so68/utils/logger"
    "log/slog"
)

func main() {
    config := &logger.Config{
        Level: slog.LevelInfo,
        Console: logger.ConsoleConfig{
            Enabled: true,
            Color:   true,
        },
        File: logger.FileConfig{
            Enabled: true,
            Path:    "logs/app.log",
            MaxSize: 100,
            MaxAge:  30,
        },
    }
    
    log, err := logger.NewLogger(config)
    if err != nil {
        panic(err)
    }
    
    log.Info("Application started", "port", 8080)
}
```

**Key Features:**
- Structured logging with slog
- File rotation with lumberjack
- Colored console output
- Configurable log levels
- Multiple output targets

### 5. WebSocket (`socket/`)

Complete WebSocket client and server hub implementation.

#### WebSocket Client (`socket/client/`)

```go
package main

import "github.com/so68/utils/socket/client"

func main() {
    ws := client.NewWebsocket("ws://localhost:8080/ws")
    
    ws.SetMessageHandler(func(message []byte) {
        fmt.Println("Received:", string(message))
    })
    
    if err := ws.Connect(); err != nil {
        log.Fatal(err)
    }
    
    ws.Send([]byte("Hello Server!"))
}
```

#### WebSocket Hub (`socket/handler/`)

```go
package main

import "github.com/so68/utils/socket/handler"

func main() {
    hub := handler.NewHub()
    
    hub.SetMessageHandler(func(connID string, message []byte) {
        fmt.Printf("Message from %s: %s\n", connID, string(message))
    })
    
    go hub.Start()
    
    // Handle WebSocket connections
    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        hub.HandleConnection(w, r)
    })
}
```

**Key Features:**
- Automatic reconnection
- Message broadcasting
- Connection management
- Event handling
- Configurable ping/pong

## Testing

The library includes comprehensive test coverage:

```bash
go test ./...
```

## Requirements

- Go 1.24.2 or later
- Dependencies are managed via `go.mod`

## Dependencies

- `github.com/gorilla/websocket` - WebSocket implementation
- `github.com/mattn/go-colorable` - Colored console output
- `gopkg.in/natefinch/lumberjack.v2` - Log file rotation

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Examples

See the `*_test.go` files for comprehensive usage examples of each module.
