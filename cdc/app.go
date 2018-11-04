// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cdc

import (
	"context"
	"fmt"
	"net/http"
	"runtime"

	"github.com/bborbe/run"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// App for streaming changes from Mariadb to Kafka
type App struct {
	CdcDatabase  string
	CdcFormat    string
	CdcHost      string
	CdcPassword  string
	CdcPort      int
	CdcTable     string
	CdcUser      string
	CdcUUID      string
	KafkaBrokers string
	KafkaTopic   string
	Port         int
}

// Validate returns an error if not all required parameter are set
func (a *App) Validate() error {
	if a.Port <= 0 {
		return errors.New("Port missing")
	}
	if a.KafkaBrokers == "" {
		return errors.New("KafkaBrokers missing")
	}
	if a.KafkaTopic == "" {
		return errors.New("KafkaTopic missing")
	}
	if a.CdcHost == "" {
		return errors.New("CdcHost missing")
	}
	if a.CdcPort <= 0 {
		return errors.New("CdcPort missing")
	}
	if a.CdcUser == "" {
		return errors.New("CdcUser missing")
	}
	if a.CdcPassword == "" {
		return errors.New("CdcPassword missing")
	}
	if a.CdcDatabase == "" {
		return errors.New("CdcDatabase missing")
	}
	if a.CdcTable == "" {
		return errors.New("CdcTable missing")
	}
	if a.CdcUUID == "" {
		return errors.New("CdcUUID missing")
	}
	if a.CdcTable == "" {
		return errors.New("CdcTable missing")
	}
	if a.CdcFormat != "JSON" && a.CdcFormat != "AVRO" {
		return errors.New("CdcFormat invalid")
	}
	return nil
}

// Run the app and blocks until error occurred or the context is canceled
func (a *App) Run(ctx context.Context) error {
	reader := &Reader{
		Dialer: &TcpDialer{
			Address: fmt.Sprintf("%s:%d", a.CdcHost, a.CdcPort),
		},
		User:     a.CdcUser,
		Password: a.CdcPassword,
		Database: a.CdcDatabase,
		Table:    a.CdcTable,
		Format:   a.CdcFormat,
		UUID:     a.CdcUUID,
	}
	sender := &Sender{
		KafkaBrokers: a.KafkaBrokers,
		KafkaTopic:   a.KafkaTopic,
	}
	ch := make(chan []byte, runtime.NumCPU())
	return run.CancelOnFirstFinish(
		ctx,
		a.runHttpServer,
		func(ctx context.Context) error {
			return reader.Read(ctx, ch)
		},
		func(ctx context.Context) error {
			return sender.Send(ctx, ch)
		},
	)
}

func (a *App) runHttpServer(ctx context.Context) error {
	router := mux.NewRouter()
	router.HandleFunc("/healthz", a.check)
	router.HandleFunc("/readiness", a.check)
	router.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.Port),
		Handler: router,
	}
	go func() {
		select {
		case <-ctx.Done():
			if err := server.Shutdown(ctx); err != nil {
				glog.Warningf("shutdown failed: %v", err)
			}
		}
	}()
	return server.ListenAndServe()
}

func (a *App) check(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(http.StatusOK)
}
