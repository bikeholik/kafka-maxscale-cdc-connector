// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cdc

import (
	"context"
	"net"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

const connectTimeout = 30 * time.Second

//go:generate counterfeiter -o ../mocks/connection.go --fake-name Connection . Connection
type Connection interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
}

type TcpDialer struct {
	Address string
}

func (t *TcpDialer) Dial(ctx context.Context) (Connection, error) {
	glog.V(2).Infof("connect to %s", t.Address)
	dialer := net.Dialer{
		Timeout: connectTimeout,
	}
	conn, err := dialer.DialContext(ctx, "tcp", t.Address)
	if err != nil {
		return nil, errors.Wrapf(err, "connect to %s failed", t.Address)
	}
	return conn, nil
}
