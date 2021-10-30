# HTTP server bootstrap

Designed for go-flags package but can be used independently.

Motivation: I tired to write the same bootstrap code again and again in my project.

Features:

- native integration with [go-flags](https://github.com/jessevdk/go-flags)
- context-aware: gracefully finished execution when context closed
- supports TLS and Auto-TLS (Let's Encrypt)


## Installation

    go get github.com/reddec/run-http-server


## Usage

```go
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	"github.com/jessevdk/go-flags"
	"github.com/reddec/run-http-server"
)

func main() {
	var server httpserver.Server
	flags.Parse(&server)

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello world"))
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	panic(server.Run(ctx))
}



```