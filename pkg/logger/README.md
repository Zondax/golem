# logger package

The logger package is intended to make unified log management in the whole app

The log may be used in 2 ways:

- Global: easy use of the global logger without the need to init or configure. 
```
package main

import (
    "github.com/zondax/golem/pkg/logger"
)

func main() {
    // Importing logger global logger is configured in info level and may be used
    logger.Info("Log info message")

    // Reconfigure global logger with config
    logger.SetGlobalConfig(logger.Config{Level: "debug"})
    logger.Info("Log debug message")
    logger.Sync()
}
```

- Local: use distinct logger for the package
```
package main

import (
    "github.com/zondax/golem/pkg/logger"
)

func main() {
    // Generate new logger with options and use it
    log := logger.NewLogger(opts ...interface{})
    log.Info("Log info message")
    log.Sync()
}
```
