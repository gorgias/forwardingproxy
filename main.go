// Copyright (C) 2018 Betalo AB - All Rights Reserved

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	var (
		flagCertPath                = flag.String("cert", "", "Filepath to certificate")
		flagKeyPath                 = flag.String("key", "", "Filepath to private key")
		flagAddr                    = flag.String("addr", "", "Server address")
		flagAuthUser                = flag.String("user", "", "Server authentication username")
		flagAuthPass                = flag.String("pass", "", "Server authentication password")
		flagDestDialTimeout         = flag.Duration("destdialtimeout", 10*time.Second, "Destination dial timeout")
		flagDestReadTimeout         = flag.Duration("destreadtimeout", 5*time.Second, "Destination read timeout")
		flagDestWriteTimeout        = flag.Duration("destwritetimeout", 5*time.Second, "Destination write timeout")
		flagClientReadTimeout       = flag.Duration("clientreadtimeout", 5*time.Second, "Client read timeout")
		flagClientWriteTimeout      = flag.Duration("clientwritetimeout", 5*time.Second, "Client write timeout")
		flagServerReadTimeout       = flag.Duration("serverreadtimeout", 30*time.Second, "Server read timeout")
		flagServerReadHeaderTimeout = flag.Duration("serverreadheadertimeout", 30*time.Second, "Server read header timeout")
		flagServerWriteTimeout      = flag.Duration("serverwritetimeout", 30*time.Second, "Server write timeout")
		flagServerIdleTimeout       = flag.Duration("serveridletimeout", 30*time.Second, "Server idle timeout")
		flagVerbose                 = flag.Bool("verbose", false, "Set log level to DEBUG")
	)

	flag.Parse()

	c := zap.NewProductionConfig()
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if *flagVerbose {
		c.Level.SetLevel(zapcore.DebugLevel)
	} else {
		c.Level.SetLevel(zapcore.ErrorLevel)
	}

	logger, err := c.Build()
	if err != nil {
		log.Fatalln("Error: failed to initiate logger")
	}
	defer logger.Sync()
	stdLogger := zap.NewStdLog(logger)

	p := &Proxy{
		ForwardingHTTPProxy: NewForwardingHTTPProxy(stdLogger),
		Logger:              logger,
		AuthUser:            *flagAuthUser,
		AuthPass:            *flagAuthPass,
		DestDialTimeout:     *flagDestDialTimeout,
		DestReadTimeout:     *flagDestReadTimeout,
		DestWriteTimeout:    *flagDestWriteTimeout,
		ClientReadTimeout:   *flagClientReadTimeout,
		ClientWriteTimeout:  *flagClientWriteTimeout,
	}

	s := &http.Server{
		Addr:              *flagAddr,
		Handler:           p,
		ErrorLog:          stdLogger,
		ReadTimeout:       *flagServerReadTimeout,
		ReadHeaderTimeout: *flagServerReadHeaderTimeout,
		WriteTimeout:      *flagServerWriteTimeout,
		IdleTimeout:       *flagServerIdleTimeout,
		TLSNextProto:      map[string]func(*http.Server, *tls.Conn, http.Handler){}, // Disable HTTP/2
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		p.Logger.Info("Server shutting down")
		if err = s.Shutdown(context.Background()); err != nil {
			p.Logger.Error("Server shutdown failed", zap.Error(err))
		}
		close(idleConnsClosed)
	}()

	p.Logger.Info("Server starting", zap.String("address", s.Addr))

	var svrErr error
	if *flagCertPath != "" && *flagKeyPath != "" {
		svrErr = s.ListenAndServeTLS(*flagCertPath, *flagKeyPath)
	} else {
		svrErr = s.ListenAndServe()
	}

	if svrErr != http.ErrServerClosed {
		p.Logger.Error("Listening for incoming connections failed", zap.Error(svrErr))
	}

	<-idleConnsClosed
	p.Logger.Info("Server stopped")
}
