# ZProfiller

## Overview

`zprofiller` is a Go package designed to facilitate the integration of Go's built-in `pprof` profiling into web applications using the Chi router. It offers a structured way to expose profiling endpoints to monitor and analyze the performance of Go applications in both development and production environments.

## Features

- **Easy Integration**: Seamlessly integrates with the Chi router.
- **Automatic Profiling Endpoints**: Automatically registers standard `pprof` endpoints.
- **Custom Configuration**: Allows customization of server timeouts and logging.

## Getting Started

### Installation

To use `zprofiller` in your project, ensure you have Go installed and your workspace is set up, then add `zprofiller` to your dependencies:

```
go get -u github.com/zondax/zprofiller
```

### Integration

Integrate `zprofiller` into your Go application:

1. **Import the Package**

   ```go
   import (
       "github.com/zondax/zprofiller"
   )
   ```

2. **Create a Config Object**

   ```go
   config := &zprofiller.Config{
       ReadTimeOut:  5 * time.Second,
       WriteTimeOut: 5 * time.Second,
       Logger:       logger.NewLogger(),
   }
   ```

3. **Instantiate zprofiller**

   ```go
   profiler := zprofiller.New(nil, config)
   ```

4. **Run the Profiler**

   ```go
   err := profiler.Run(":9999")
   if err != nil {
       log.Fatalf("Failed to start profiler: %v", err)
   }
   ```

### Usage

Access the profiling endpoints at `http://localhost:<port>/debug/pprof/`, where `<port>` is the port you specified.

## Viewing pprof Results on the Web

### Accessing pprof via Web Browser

Navigate to:

```
http://localhost:<port>/debug/pprof/
```

This index page links to profiles like Heap, Goroutine, Threadcreate, Block, and Mutex.

### Visualizing Profiles

Use tools like `Go Tool Pprof` or `Graphviz` for deeper analysis:

- **Go Tool Pprof**:

  ```
  go tool pprof -http=:8081 http://localhost:<port>/debug/pprof/profile
  ```

  This command downloads the CPU profile data from your application and opens it in an interactive web interface on `http://localhost:8081`.

- **Graphviz**:

  ```
  sudo apt-get install graphviz
  go tool pprof -http=:8081 --graph http://localhost:<port>/debug/pprof/profile
  ```

### Online Tools and Extensions

Consider using online tools or browser extensions like **pprof++** for Chrome for in-browser visualization of pprof data.

## Performance Considerations

When integrating profiling tools such as `pprof` into your application, it is essential to consider the potential impact on performance:

- **Resource Usage**: Profiling operations can consume significant CPU and memory resources, particularly when capturing and analyzing high-frequency data such as CPU profiles.
- **Production Use**: While `pprof` can be invaluable for diagnosing issues in production, it should be enabled selectively. Consider using environment variables or configuration files to control access to profiling endpoints.
- **Sampling Rate**: Adjust the sampling rate of profiles according to the performance impact and the level of detail required. Lower rates can reduce overhead but may miss critical details.
- **Security**: Exposing profiling information can introduce security risks. Ensure that profiling endpoints are protected with authentication mechanisms and are only accessible by authorized personnel.
- **Impact Measurement**: Continuously monitor the impact of enabling profiling on your system’s response times and resource usage. Disable profiling when not needed to avoid unnecessary overhead.

## Security Considerations

Ensure that access to profiling endpoints is secured, especially in production environments, to protect sensitive application data.
