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
- **OpenTelemetry Integration**: Configurable observability with tracing and metrics for database operations.
- **Mocks**:  Provides mock implementations for DBConnector and ZDatabase interfaces, making it straightforward to unit test the package components without setting up actual database connections

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
// OpenTelemetry configuration
OpenTelemetry: zdbconfig.OpenTelemetryConfig{
    Enabled: true,
    IncludeQueryParameters: true,
    QueryFormatter: zdbconfig.QueryFormatterLower,
    DefaultAttributes: map[string]string{
        "service": "my-service",
    },
},
}

    db, err := zdatabase.NewInstance(zdbConnector.DBTypeClickhouse, config)
    if err != nil {
        panic(err)
    }

    // Perform operations
}
\`\`\`

## Configuration

The `zdatabase` package can be configured through the `zdbconfig` package. OpenTelemetry instrumentation supports query parameter inclusion, query formatting (upper/lower/none/default), custom attributes, and metrics collection. Refer to [Configuration Documentation](docs/configuration.md) for more details.

## Support

- clickhouse
- postgresql
- cloud sql postgres

## Cloud SQL PostgreSQL Connector

### Configuration

```go
config := &zdbconfig.Config{
    ConnectionParams: zdbconfig.ConnectionParams{
        User: "service-account@project-id.iam", // For IAM auth
        Name: "database-name",
        CloudSQL: zdbconfig.CloudSQLConfig{
            Enabled:      true,
            InstanceName: "project:region:instance",
            UseIAMAuth:   true,  // Recommended for production
            UsePrivateIP: true,  // For VPC connections
        },
    },
    LogConfig: zdbconfig.LogConfig{
        LogLevel: "info",
    },
}

db, err := zdb.NewInstance(zdbconnector.DBTypeCloudSQLPostgres, config)
```

### Authentication Options

**IAM Authentication (Recommended):**
```go
CloudSQL: zdbconfig.CloudSQLConfig{
    UseIAMAuth: true,
    // No password needed
}
```

**Password Authentication:**
```go
ConnectionParams: zdbconfig.ConnectionParams{
    Password: "your-password",
},
CloudSQL: zdbconfig.CloudSQLConfig{
    UseIAMAuth: false,
}
```

### Instance Name Format
Always use: `"project-id:region:instance-name"`