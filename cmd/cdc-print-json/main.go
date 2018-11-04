package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/bborbe/kafka-maxscale-cdc-connector/cdc"
	"github.com/golang/glog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())

	_ = flag.Set("logtostderr", "true")
	flag.Parse()
	ctx := contextWithSig(context.Background())

	cdcClient := &cdc.Reader{
		Address:  "localhost:4001",
		User:     "cdcuser",
		Password: "cdc",
		Database: "test",
		Table:    "names",
		Format:   "JSON",
	}

	ch := make(chan map[string]interface{}, runtime.NumCPU())
	go func() {
		for line := range ch {
			err := json.NewEncoder(os.Stdout).Encode(line)
			if err != nil {
				glog.Exitf("encode json failed")
			}
		}
	}()

	err := cdcClient.Read(ctx, ch)
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
