# Task Metrics using Prometheus in Go

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Code Structure](#code-structure)
- [Metrics Collectors](#metrics-collectors)
  - [Counter](#counter)
  - [Gauge](#gauge)
  - [Histogram](#histogram)
- [Usage](#usage)
  - [Starting the Server](#starting-the-server)
  - [Registering Metrics](#registering-metrics)
  - [Updating Metrics](#updating-metrics)

## Introduction

This project provides a comprehensive metrics collection and reporting framework integrated with Prometheus for Go applications.

## Features

- Different types of metrics collectors including Counter, Gauge, and Histogram.
- Built-in Prometheus server.
- Strong typing to prevent metrics misuse.
- Error handling and logging integrated with Uber's Zap library.
- Support for custom labels.
- Thread-safe metric updates with read-write mutexes.
- Auto-registering of metrics based on type.

## Code Structure

- `/metrics/collectors/`: Contains the implementation of various metrics collectors (Counter, Gauge, Histogram).
- `/metrics/handler.go`: Contains the MetricHandler interface and taskMetrics `UpdateMetric` method for updating metrics.
- `/metrics/prometheus.go`: Defines the Prometheus server
- `/metrics/register.go`: Responsible for registering metrics with the Prometheus server.

## Metrics Collectors

### Counter

- **File**: `/metrics/collectors/counter.go`
- **Methods**: `Update`
- **Usage**: Counters are cumulative metrics that can only increase.

### Gauge

- **File**: `/metrics/collectors/gauge.go`
- **Methods**: `Update`
- **Usage**: Gauges are metrics that can arbitrarily go up and down.

### Histogram

- **File**: `/metrics/collectors/histogram.go`
- **Methods**: `Update`
- **Usage**: Histograms count observations (like request durations or response sizes) and place them in configurable buckets.

## Usage

### Starting the Server

```go
metricsServer := metrics.NewTaskMetrics("/metrics", "9090")
err := metricsServer.Start()
if err != nil {
    log.Fatal(err)
}
```

### Registering Metrics

```go
// Without labels
err = metricsServer.RegisterMetric("my_counter", "This is a counter metric", nil, &collectors.Counter{})

// With labels
err = metricsServer.RegisterMetric("my_counter_with_labels", "This is a counter metric with labels", []string{"label1", "label2"}, &collectors.Counter{})

if err != nil {
log.Fatal(err)
}
```

### Updating Metrics

```go
// Without labels
metricsServer.UpdateMetric("my_counter", 1)

// With labels
metricsServer.UpdateMetric("my_counter_with_labels", 1, "label1_value", "label2_value")
```