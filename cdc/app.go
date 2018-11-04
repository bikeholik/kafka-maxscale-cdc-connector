package cdc

import (
	"context"
	"fmt"
	"net/http"
	"github.com/bborbe/run"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"runtime"
)

type App struct {
	CdcHost      string
	CdcPassword  string
	CdcDatabase  string
	CdcTable     string
	CdcPort      int
	CdcUser      string
	KafkaBrokers string
	KafkaTopic   string
	Port         int
}

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
	return nil
}

func (a *App) Run(ctx context.Context) error {
	reader := &Reader{
		Address:  fmt.Sprintf("%s:%d", a.CdcHost, a.CdcPort),
		User:     a.CdcUser,
		Password: a.CdcPassword,
		Database: a.CdcDatabase,
		Table:    a.CdcTable,
		Format:   "JSON",
	}
	sender := &Sender{
		KafkaBrokers: a.KafkaBrokers,
		KafkaTopic:   a.KafkaTopic,
	}
	ch := make(chan map[string]interface{}, runtime.NumCPU())
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
			server.Shutdown(ctx)
		}
	}()
	return server.ListenAndServe()
}

func (a *App) check(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(http.StatusOK)
}
