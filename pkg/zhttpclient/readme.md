# zhttpclient Package

## Overview

The `zhttpclient` package serves as an abstraction layer for http requests within Golang applications. Using Resty as a foundation, it provides convenient
abstractions and backoff functions for common http requests.

## Table of Contents

1. [Features](#features)
2. [Installation](#installation)
3. [Usage](#usage)

## Features

- **Convenience Methods**: Convenient methods for http operations.
- **Retry Mechanism**: Automatic retries depending on configurable conditions.
- **Mocks**: Provides mock implementations for the ZHTTPClient and ZRequest interface.

## Installation

```bash
go get github.com/zondax/golem/pkg/zhttpclient
```

## Usage

Here's a quick example demonstrating how to use it.

```go
import (
    "github.com/zondax/golem/pkg/zhttpclient"
)

func main() {
    config := &zhttpclient.Config{
        Timeout: ..,
        TLSConfig: ...,
        BaseClient: ..., // a pre-configured http.Client
    }

    client, err := zhttpclient.New(config)
    if err != nil {
        panic(err)
    }


    // all requests with this client will use it
    retry := &RetryPolicy{
        MaxAttempts: ..., // max number of retries
        WaitBeforeRetry: ..., // the minimum default wait before retry
        MaxWaitBeforeRetry: ..., // the maximum cap for the wait before retry
    }

    // The default backoff policy is Exponential Jitter provided by resty

    retry.SetLinearBackoff(duration)
    // or
    retry.SetExponentialBackoff(duration)

    // top-level retry policy
    client.SetRetryPolicy(retry)

    req := client.NewRequest().SetURL(srv.URL).SetHeaders(headers)

    // GET
    resp,err := req.SetQueryParams(getParams).
    		SetRetryPolicy(&zhttpclient.RetryPolicy{}). // override client retry policy
      	Get(ctx)

    // POST
    resp,err := req.SetBody(body).Post(ctx)
    fmt.Println(resp.Code,string(resp.Body))

    // AUTO-decode response
    resp,err := req.SetBody(body).SetResponse(&MyRespStruct{}).SetError(&MyErrStruct{}).Post(ctx)
    if resp.Response != nil{
    	parsedResp := resp.Response.(*MyRespStruct)
    }
    if resp.Error != nil{
    	parsedErr := resp.Error.(*MyErrStruct)
    }

    // Do raw request
    req, _ := http.NewRequest(method, URL, body)
    resp,err := client.Do(ctx,req)
    fmt.Println(string(resp.Body))

}
```
