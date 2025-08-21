
# 🥯 bunslog

A simple logger hook for [bun](https://github.com/uptrace/bun) ORM using Go's `slog` package. It logs SQL queries, errors, and slow queries with flexible configuration. 🚀

## ✨ Features

- 🐞 Log all queries
- 🎚️ Custom log levels for queries, errors, and slow queries
- ⚙️ Environment variable configuration
- 🔗 Integrates with Go's `slog` logger

## 📦 Installation

```sh
go get github.com/XanderD99/bunslog
```

## 📝 Usage Example

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "time"
    "github.com/uptrace/bun"
    "github.com/XanderD99/bunslog"
)

func main() {
    db := bun.NewDB(/* ... */)
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    hook := bunslog.NewQueryHook(
        bunslog.WithLogger(logger),
        bunslog.WithLogSlow(100 * time.Millisecond), // log queries slower than 100ms
    )
    db.AddQueryHook(hook)
    // ... your queries ...
}
```

## 🌍 Environment Configuration

You can control logging via environment variables:

- 📴 `BUNDEBUG=0` disables logging
- ✅ `BUNDEBUG=1` enables logging

## 📄 License

MIT 🎉
