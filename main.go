// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	flag "github.com/bborbe/flagenv"
	"github.com/bborbe/kafka-maxscale-cdc-connector/cdc"
	"github.com/golang/glog"
	"github.com/google/uuid"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())

	app := &cdc.App{}
	flag.IntVar(&app.Port, "port", 9001, "port to listen")
	flag.StringVar(&app.CdcHost, "cdc-host", "", "cdc host")
	flag.IntVar(&app.CdcPort, "cdc-port", 4001, "cdc port")
	flag.StringVar(&app.CdcUser, "cdc-user", "", "cdc user")
	flag.StringVar(&app.CdcPassword, "cdc-password", "", "cdc password")
	flag.StringVar(&app.CdcDatabase, "cdc-database", "", "cdc database")
	flag.StringVar(&app.CdcTable, "cdc-table", "", "cdc table")
	flag.StringVar(&app.CdcUUID, "cdc-uuid", uuid.New().String(), "cdc client identifier uuid")
	flag.StringVar(&app.CdcFormat, "cdc-format", "JSON", "cdc output format (JSON|AVRO)")
	flag.StringVar(&app.KafkaBrokers, "kafka-brokers", "", "kafka brokers")
	flag.StringVar(&app.KafkaTopic, "kafka-topic", "", "kafka topic")

	_ = flag.Set("logtostderr", "true")
	flag.Parse()

	glog.V(0).Infof("Parameter CdcHost: %s", app.CdcHost)
	glog.V(0).Infof("Parameter CdcPort: %d", app.CdcPort)
	glog.V(0).Infof("Parameter CdcUser: %s", app.CdcUser)
	glog.V(0).Infof("Parameter CdcPassword-Length: %d", len(app.CdcPassword))
	glog.V(0).Infof("Parameter CdcDatabase: %s", app.CdcDatabase)
	glog.V(0).Infof("Parameter CdcTable: %s", app.CdcTable)
	glog.V(0).Infof("Parameter CdcUUID: %s", app.CdcUUID)
	glog.V(0).Infof("Parameter CdcFormat: %s", app.CdcFormat)
	glog.V(0).Infof("Parameter KafkaBrokers: %s", app.KafkaBrokers)
	glog.V(0).Infof("Parameter KafkaTopic: %s", app.KafkaTopic)
	glog.V(0).Infof("Parameter Port: %d", app.Port)

	ctx := contextWithSig(context.Background())

	glog.V(0).Infof("app started")
	if err := app.Run(ctx); err != nil {
		glog.Exitf("app failed: %+v", err)
	}
	glog.V(0).Infof("app finished")
}

func contextWithSig(ctx context.Context) context.Context {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	go func() {
		defer cancel()

		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-signalCh:
		case <-ctx.Done():
		}
	}()

	return ctxWithCancel
}
