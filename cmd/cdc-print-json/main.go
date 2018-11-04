// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Sample of read and print cdc messages
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/bborbe/kafka-maxscale-cdc-connector/cdc"
	"github.com/golang/glog"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())

	_ = flag.Set("logtostderr", "true")
	flag.Parse()
	ctx := contextWithSig(context.Background())

	cdcClient := &cdc.MaxscaleReader{
		Dialer: &cdc.TcpDialer{
			Address: "localhost:4001",
		},
		User:     "cdcuser",
		Password: "cdc",
		Database: "test",
		Table:    "names",
		Format:   "JSON",
	}

	ch := make(chan []byte, runtime.NumCPU())
	go func() {
		for line := range ch {
			os.Stdout.Write(line)
		}
	}()

	err := cdcClient.Read(ctx, nil, ch)
	if err != nil {
		glog.Exitf("%+v", err)
	}
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
		glog.V(1).Infof("shutdown now ...")
	}()

	return ctxWithCancel
}
