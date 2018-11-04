// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cdc

import (
	"context"
	"runtime"
	"sync"

	"github.com/bborbe/run"
)

// Reader interface for the Streamer
//go:generate counterfeiter -o ../mocks/reader.go --fake-name Reader . Reader
type Reader interface {
	// Read changes and send them to the given channel
	Read(ctx context.Context, gtid *GTID, ch chan<- []byte) error
}

// Sender interface for the Streamer
//go:generate counterfeiter -o ../mocks/sender.go --fake-name Sender . Sender
type Sender interface {
	Send(ctx context.Context, ch <-chan []byte) error
}

// Streamer coordinates read and send of CDC records
type Streamer struct {
	GTID   *GTID
	Reader Reader
	Sender Sender
}

// Run read and send of CDC records
func (s *Streamer) Run(ctx context.Context) error {
	ch := make(chan []byte, runtime.NumCPU())
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		wg.Wait()
		close(ch)
	}()
	return run.CancelOnFirstFinish(
		ctx,
		func(ctx context.Context) error {
			defer wg.Done()
			return s.Reader.Read(ctx, s.GTID, ch)
		},
		func(ctx context.Context) error {
			defer wg.Done()
			return s.Sender.Send(ctx, ch)
		},
	)
}
