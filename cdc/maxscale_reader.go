// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cdc

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// Dialer interface for open a new connection
//go:generate counterfeiter -o ../mocks/dialer.go --fake-name Dialer . Dialer
type Dialer interface {
	Dial(ctx context.Context) (Connection, error)
}

// MaxscaleReader of CDC messages from Maxscale
type MaxscaleReader struct {
	Dialer   Dialer
	User     string
	Password string
	UUID     string
	Format   string // JSON or AVRO
	Database string
	Table    string
	Version  string
}

// Read all cdc and send them to the given channel
// https://mariadb.com/resources/blog/how-to-stream-change-data-through-mariadb-maxscale-using-cdc-api/
func (r *MaxscaleReader) Read(ctx context.Context, gtid *GTID, ch chan<- []byte) error {

	conn, err := r.Dialer.Dial(ctx)
	if err != nil {
		return errors.Wrapf(err, "connect to cdc failed")
	}
	defer conn.Close()

	err = r.writeAuth(conn)
	if err != nil {
		return err
	}
	err = r.expectResponse(conn, []byte("OK"))
	if err != nil {
		return err
	}
	glog.V(1).Infof("login successful")

	_, err = fmt.Fprintf(conn, "REGISTER UUID=%s, TYPE=%s", r.UUID, r.Format)
	if err != nil {
		return errors.Wrapf(err, "register with uuid: %s and type: %s failed", r.UUID, r.Format)
	}
	err = r.expectResponse(conn, []byte("OK"))
	if err != nil {
		return err
	}
	glog.V(1).Infof("register with uuid: %s and type: %s successful", r.UUID, r.Format)

	_, err = conn.Write(r.buildRequestCommand(gtid))
	if err != nil {
		return errors.Wrap(err, "write request to connection failed")
	}

	errs := make(chan error)
	glog.V(1).Infof("start streaming of %s %s %s %s", r.Database, r.Table, r.Version, gtid)
	go func() {
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadBytes('\n')
			if err == io.EOF {
				glog.V(1).Infof("connection closed")
				errs <- nil
				return
			}
			if err != nil {
				errs <- errors.Wrap(err, "read line failed")
				return
			}
			if startsWith(line, []byte("ERR")) {
				errs <- errors.Errorf("got error: %s", string(line))
				return
			}
			if glog.V(4) {
				glog.Infof("read %s", string(line))
			}
			select {
			case ch <- line:
			case <-ctx.Done():
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errs:
		return err
	}
}

// REQUEST-DATA DATABASE.TABLE[.VERSION] [GTID]
func (r *MaxscaleReader) buildRequestCommand(gtid *GTID) []byte {
	buf := bytes.NewBufferString("REQUEST-DATA ")
	_, _ = fmt.Fprintf(buf, "%s.%s", r.Database, r.Table)
	if len(r.Version) > 0 {
		_, _ = fmt.Fprintf(buf, ".%s", r.Version)
	}
	if gtid != nil {
		_, _ = fmt.Fprintf(buf, " %s", gtid.String())
	}
	return buf.Bytes()
}

func (r *MaxscaleReader) expectResponse(conn io.Reader, expectedResponse []byte) error {
	buf, err := r.read(conn)
	if err != nil {
		return err
	}
	if !startsWith(buf, expectedResponse) {
		return errors.New("login failed")
	}
	return nil
}

func (r *MaxscaleReader) read(conn io.Reader) ([]byte, error) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, errors.Wrap(err, "read failed")
	}
	return buf[0:n], nil
}

func (r *MaxscaleReader) writeAuth(conn io.Writer) error {
	h := sha1.New()
	io.WriteString(h, r.Password)

	encoder := hex.NewEncoder(conn)
	_, err := encoder.Write([]byte(fmt.Sprintf("%s:", r.User)))
	if err != nil {
		return errors.Wrap(err, "hex encode failed")
	}
	_, err = encoder.Write(h.Sum(nil))
	if err != nil {
		return errors.Wrap(err, "write failed")
	}
	return nil
}

func startsWith(line []byte, prefix []byte) bool {
	if len(line) < len(prefix) {
		return false
	}
	return bytes.Equal(line[0:len(prefix)], prefix)
}
