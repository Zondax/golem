# ZRouter package

ZRouter is a Golang routing library built on the robust foundation of the chi router.

## Table of Contents

- [Features](#features)
- [Getting Started](#getting-started)
    - [Installation](#installation)
- [Usage](#usage)
- [Routing](#routing)
- [Middleware](#middleware)
- [Adapters](#adapters)
- [Custom Configurations](#custom-configurations)
- [Monitoring and Logging](#monitoring-and-logging)
- [Advanced Topics](#advanced-topics)
- [Examples](#examples)
- [Conclusion](#conclusion)

## Features

- **Intuitive Routing Interface**: Define your routes with ease using all major HTTP methods.
- **Middleware Chaining**: Introduce layers of middleware to your HTTP requests and responses.
- **Flexible Adapters**: Seamlessly integrate with the `chi` router's context.
- **Enhanced Monitoring**: Integrated metrics server and structured logging for in-depth observability.
- **Customizable Settings**: Adjust server configurations like timeouts to suit your needs.

## Getting Started

### Installation

To incorporate ZRouter into your project:

```bash
go get github.com/zondax/golem/pkg/zrouter
```

## Usage

Crafting a web service using ZRouter:

```go
import "github.com/zondax/golem/pkg/zrouter"

config := &zrouter.Config{ReadTimeOut: 10 * time.Second, WriteTimeOut: 10 * time.Second}
router := zrouter.New("ServiceName", metricServer, config)

router.Use(middlewareLogic)

groupedRoutes := router.Group("/grouped")
groupedRoutes.GET("/{param}", handlerFunction)
```

or 

```go
import "github.com/zondax/golem/pkg/zrouter"

func main() {
router := zrouter.New("MyService", metricServer, nil)

router.GET("/endpoint", func(ctx zrouter.Context) (domain.ServiceResponse, error) {
// Handler implementation
})

router.Run()
}
```

## Routing

For dynamic URL parts, utilize the chi style, e.g., /entities/{entityID}.

## Middleware

Add pre- and post-processing steps to your routes. Chain multiple middlewares for enhanced functionality.

### Default Middlewares

ZRouter sets certain default middlewares:
- `ErrorHandlerMiddleware`: Handles errors, translating them to a standard response.
- `RequestID()`: Appends a unique request ID to each request.
- `RequestMetrics()`: Logs metrics related to requests, responses, and other interactions.

#### Default Registered Metrics

When using the `RequestMetrics()` middleware, the following metrics are registered by default:
- **Request Count**: Tracks the number of received requests.
- **Response Time**: Measures the time taken to process a request and send a response.
- **Error Count**: Monitors the number of errors thrown during request processing.

## Adapters

Use `chiContextAdapter` for translating the `chi` router's context to ZRouter's.

## Custom Configurations

Specify server behavior with `Config`. Use default settings or customize as needed.

Default settings: 
- `ReadTimeOut`: 240000 milliseconds.
- `WriteTimeOut`: 240000 milliseconds.
  Override these defaults by providing values during initialization.

## Response Standards

### ServiceResponse

When handling responses, ZRouter provides a standardized way to return them using `ServiceResponse`, which includes status, headers, and body.

**Example**:

```go
func MyHandler(ctx Context) (domain.ServiceResponse, error) {
    data := map[string]string{"message": "Hello, World!"}
    return domain.NewServiceResponse(http.StatusOK, data), nil
}
```

### Handling Headers

With `ServiceResponse`, you can easily set custom headers for your responses:

```go
func MyHandler(ctx Context) (domain.ServiceResponse, error) {
  headers := make(http.Header)
  headers.Set("X-Custom-Header", "My Value")
  
  data := map[string]string{"message": "Hello, World!"}
  response := domain.NewServiceResponseWithHeader(http.StatusOK, data, headers)
  return response, nil
}
```
### Error Handling

Whenever you return an error, ZRouter translates it to a structured error response, maintaining consistency across your services.

**Example**:

```go
func MyHandler(ctx Context) (domain.ServiceResponse, error) {
    return nil, domain.NewAPIErrorResponse(http.StatusNotFound, "not_found", "message")
}
```

## Context in ZRouter

The `Context` is an essential part of ZRouter, providing a consistent interface to interact with the HTTP request and offering helper methods to streamline handler operations. This abstraction ensures that, as your router's needs evolve, the core interface to access request information remains consistent.

### Functions and Usage:

1. **Request**:

   Retrieve the raw `*http.Request` from the context:

    ```go
    req := ctx.Request()
    ```

2. **BindJSON**:

   Decode a JSON request body directly into a provided object:

    ```go
    var myData MyStruct
    err := ctx.BindJSON(&myData)
    ```

3. **Header**:

   Set an HTTP header for the response:

    ```go
    ctx.Header("X-Custom-Header", "Custom Value")
    ```

4. **Param**:

   Get URL parameters (path variables):

    ```go
    userID := ctx.Param("userID")
    ```

5. **Query**:

   Retrieve a query parameter from the URL:

    ```go
    sortBy := ctx.Query("sortBy")
    ```

6. **DefaultQuery**:

   Retrieve a query parameter from the URL, but return a default value if it's not present:

    ```go
    order := ctx.DefaultQuery("order", "asc")
    ```

### Adapting to chi:

Behind the scenes, ZRouter leverages the powerful `chi` router. The `chiContextAdapter` translates the chi context to ZRouter's, ensuring that you get the benefits of chi's speed and power with ZRouter's simplified and consistent interface.

## Monitoring and Logging

Monitor request metrics and employ structured logging for in-depth insights.

## Advanced Topics

- **Route Grouping**: Consolidate routes under specific prefixes using `Group()`.
- **NotFound Handling**: Specify custom logic for unmatched routes.
- **Route Tracking**: Fetch a structured list of all registered routes.
