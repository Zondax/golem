# zdatabase Package

## Overview
The `zdatabase` package serves as an abstraction layer for database interactions within Golang applications. Using GORM as a foundation, it provides a more straightforward and extensible interface for database operations.

## Table of Contents
1. [Features](#features)
2. [Installation](#installation)
3. [Usage](#usage)
4. [Configuration](#configuration)

## Features
- **Common Interface**: A unified API for different database operations.
- **Extensible**: Connector mechanism to add support for additional database types.
- **Robustness**: In-built logging and error-handling features.
- **Retry Mechanism**: Automatic retries for failed database connections.
- **zdb mock**: zdb mock for testing 

## Installation
\`\`\`bash
go get github.com/zondax/golem/pkg/zdatabase
\`\`\`

## Usage

Here's a quick example demonstrating how to create a new database instance.

\`\`\`go
import (
"github.com/zondax/golem/pkg/zdatabase"
"github.com/zondax/golem/pkg/zdatabase/zdbconfig"
)

func main() {
config := &zdbconfig.Config{
// Connection settings
}

    db, err := zdatabase.NewInstance(zdbConnector.DBTypeClickhouse, config)
    if err != nil {
        panic(err)
    }

    // Perform operations
}
\`\`\`

## Configuration

The `zdatabase` package can be configured through the `zdbconfig` package. Refer to [Configuration Documentation](docs/configuration.md) for more details.
