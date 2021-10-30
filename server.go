package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/crypto/acme/autocert"
)

type Server struct {
	Bind             string        `long:"bind" env:"BIND" description:"Binding address" default:"127.0.0.1:8080"`
	GracefulShutdown time.Duration `long:"graceful-shutdown" env:"GRACEFUL_SHUTDOWN" description:"Timeout for HTTP server graceful shutdown" default:"10s"`
	TLS              bool          `long:"tls" env:"TLS" description:"Enable TLS server"`
	TLSCert          string        `long:"tls-cert" env:"TLS_CERT" description:"Path to TLS certificate" default:"server.crt"`
	TLSKey           string        `long:"tls-key" env:"TLS_KEY" description:"Path to TLS private key" default:"server.key"`
	AutoTLS          bool          `long:"auto-tls" env:"AUTO_TLS" description:"Enable automatic TLS certificates by Let's Encrypt. Forces bind to :443"`
	AutoTLSDomains   []string      `long:"auto-tls-domains" env:"AUTO_TLS_DOMAINS" description:"Restrict automatic TLS certificates to specified domains"`
	AutoTLSCache     string        `long:"auto-tls-cache" env:"AUTO_TLS_CACHE" description:"Directory to store obtained certificates" default:".certs"`
}

func (config Server) Run(ctx context.Context) error {
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

	})
	return config.Serve(ctx, &http.Server{
		Addr:    config.Bind,
		Handler: http.DefaultServeMux,
	})
}

func (config Server) Serve(global context.Context, server *http.Server) error {
	ctx, cancel := context.WithCancel(global)
	defer cancel()

	done := make(chan error)

	go func() {
		defer close(done)
		<-ctx.Done()
		graceful, cancel := context.WithTimeout(context.Background(), config.GracefulShutdown)
		defer cancel()
		done <- server.Shutdown(graceful)
	}()

	var res *multierror.Error
	err := config.runServer(server)
	cancel()
	shutDown := <-done
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		res = multierror.Append(res, err)
	}
	if shutDown != nil && !errors.Is(shutDown, http.ErrServerClosed) {
		res = multierror.Append(res, shutDown)
	}
	return res.ErrorOrNil()
}

func (config Server) runServer(server *http.Server) error {
	switch {
	case config.AutoTLS:
		manager := &autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache(config.AutoTLSCache),
		}
		if len(config.AutoTLSDomains) > 0 {
			manager.HostPolicy = autocert.HostWhitelist(config.AutoTLSDomains...)
		}
		l := manager.Listener()
		defer l.Close()

		return server.Serve(l)
	case config.TLS:
		return server.ListenAndServeTLS(config.TLSCert, config.TLSKey)
	default:
		return server.ListenAndServe()
	}
}
