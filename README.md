# Go Logger

Common logging library for go

example usage:

```go
package main

import "github.com/Aibier/go-logger"

func main() {
    cfg := logger.Config{Log: "Dev"}
    log, err := logger.NewZapLogger(cfg)

    logger.With("key", "value").Info("Messahe")
}

```

# go-logger
