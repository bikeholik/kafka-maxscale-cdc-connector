// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cdc_test

import (
	"context"
	"fmt"
	"io"

	"github.com/bborbe/kafka-maxscale-cdc-connector/cdc"
	"github.com/bborbe/kafka-maxscale-cdc-connector/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

var _ = Describe("Reader", func() {
	var reader *cdc.Reader
	var dialer *mocks.Dialer
	var conn *mocks.Connection

	BeforeEach(func() {
		dialer = &mocks.Dialer{}
		conn = &mocks.Connection{}
		dialer.DialReturns(conn, nil)
		reader = &cdc.Reader{
			Dialer:   dialer,
			User:     "cdcuser",
			Password: "cdc",
			UUID:     "0f672312-e02a-11e8-8c13-cf8f48795343",
			Format:   "JSON",
			Database: "mydb",
			Table:    "mytable",
		}
	})

	It("returns error if dial fails", func() {
		dialer.DialReturns(nil, errors.New("banana"))
		err := reader.Read(context.Background(), make(chan []byte))
		Expect(err).NotTo(BeNil())
		Expect(conn.WriteCallCount()).To(Equal(0))
		Expect(conn.ReadCallCount()).To(Equal(0))
	})

	It("writes auth to connection", func() {
		err := reader.Read(context.Background(), make(chan []byte))
		Expect(err).NotTo(BeNil())
		Expect(conn.WriteCallCount()).To(Equal(2))
		Expect(conn.ReadCallCount()).To(Equal(1))
		Expect(string(conn.WriteArgsForCall(0))).To(Equal("636463757365723a"))
		Expect(string(conn.WriteArgsForCall(1))).To(Equal("11fb6d5a105a66c85408b8005c461d53818b736e"))
	})

	It("returns error if writes auth failed", func() {
		conn.WriteReturns(0, errors.New("write banana"))
		err := reader.Read(context.Background(), make(chan []byte))
		Expect(err).NotTo(BeNil())
		Expect(conn.WriteCallCount()).To(Equal(1))
		Expect(conn.ReadCallCount()).To(Equal(0))
	})

	It("returns error if read failed", func() {
		conn.ReadReturns(0, errors.New("read banana"))
		err := reader.Read(context.Background(), make(chan []byte))
		Expect(err).NotTo(BeNil())
		Expect(conn.WriteCallCount()).To(Equal(2))
		Expect(conn.ReadCallCount()).To(Equal(1))
		Expect(string(conn.WriteArgsForCall(0))).To(Equal("636463757365723a"))
		Expect(string(conn.WriteArgsForCall(1))).To(Equal("11fb6d5a105a66c85408b8005c461d53818b736e"))
	})

	It("returns error if read != OK", func() {
		conn.ReadStub = func(bytes []byte) (int, error) {
			n := copy(bytes[:], "ERR banana")
			return n, nil
		}
		err := reader.Read(context.Background(), make(chan []byte))
		Expect(err).NotTo(BeNil())
		Expect(conn.WriteCallCount()).To(Equal(2))
		Expect(conn.ReadCallCount()).To(Equal(1))
		Expect(string(conn.WriteArgsForCall(0))).To(Equal("636463757365723a"))
		Expect(string(conn.WriteArgsForCall(1))).To(Equal("11fb6d5a105a66c85408b8005c461d53818b736e"))
	})

	It("send register after successful auth", func() {
		readCounter := 0
		conn.ReadStub = func(bytes []byte) (int, error) {
			readCounter++
			if readCounter == 1 {
				n := copy(bytes[:], "OK\n")
				return n, nil
			}
			return 0, errors.New("read banana")
		}
		err := reader.Read(context.Background(), make(chan []byte))
		Expect(err).NotTo(BeNil())
		Expect(conn.WriteCallCount()).To(Equal(3))
		Expect(string(conn.WriteArgsForCall(0))).To(Equal("636463757365723a"))
		Expect(string(conn.WriteArgsForCall(1))).To(Equal("11fb6d5a105a66c85408b8005c461d53818b736e"))
		Expect(string(conn.WriteArgsForCall(2))).To(Equal("REGISTER UUID=0f672312-e02a-11e8-8c13-cf8f48795343, TYPE=JSON"))
	})

	It("returns error if send register failed", func() {
		conn.ReadStub = func(bytes []byte) (int, error) {
			n := copy(bytes[:], "OK\n")
			return n, nil
		}
		conn.WriteReturnsOnCall(2, 0, errors.New("write banana"))
		err := reader.Read(context.Background(), make(chan []byte))
		Expect(err).NotTo(BeNil())
		Expect(conn.WriteCallCount()).To(Equal(3))
		Expect(string(conn.WriteArgsForCall(0))).To(Equal("636463757365723a"))
		Expect(string(conn.WriteArgsForCall(1))).To(Equal("11fb6d5a105a66c85408b8005c461d53818b736e"))
		Expect(string(conn.WriteArgsForCall(2))).To(Equal("REGISTER UUID=0f672312-e02a-11e8-8c13-cf8f48795343, TYPE=JSON"))
	})

	It("returns error if send register returns not OK", func() {
		readCounter := 0
		conn.ReadStub = func(bytes []byte) (int, error) {
			readCounter++
			if readCounter == 1 {
				n := copy(bytes[:], "OK\n")
				return n, nil
			}
			if readCounter == 2 {
				n := copy(bytes[:], "ERR\n")
				return n, nil
			}
			return 0, errors.New("read banana")
		}
		err := reader.Read(context.Background(), make(chan []byte))
		Expect(err).NotTo(BeNil())
		Expect(conn.ReadCallCount()).To(Equal(2))
		Expect(conn.WriteCallCount()).To(Equal(3))
		Expect(string(conn.WriteArgsForCall(0))).To(Equal("636463757365723a"))
		Expect(string(conn.WriteArgsForCall(1))).To(Equal("11fb6d5a105a66c85408b8005c461d53818b736e"))
		Expect(string(conn.WriteArgsForCall(2))).To(Equal("REGISTER UUID=0f672312-e02a-11e8-8c13-cf8f48795343, TYPE=JSON"))
	})

	It("send request if register returns OK", func() {
		readCounter := 0
		conn.ReadStub = func(bytes []byte) (int, error) {
			readCounter++
			if readCounter == 1 || readCounter == 2 {
				n := copy(bytes[:], "OK\n")
				return n, nil
			}
			return 0, errors.New("read banana")
		}
		err := reader.Read(context.Background(), make(chan []byte))
		Expect(err).NotTo(BeNil())
		Expect(conn.ReadCallCount()).To(Equal(3))
		Expect(conn.WriteCallCount()).To(Equal(4))
		Expect(string(conn.WriteArgsForCall(0))).To(Equal("636463757365723a"))
		Expect(string(conn.WriteArgsForCall(1))).To(Equal("11fb6d5a105a66c85408b8005c461d53818b736e"))
		Expect(string(conn.WriteArgsForCall(2))).To(Equal("REGISTER UUID=0f672312-e02a-11e8-8c13-cf8f48795343, TYPE=JSON"))
		Expect(string(conn.WriteArgsForCall(3))).To(Equal("REQUEST-DATA mydb.mytable"))
	})

	It("returns error if request returns ERR", func() {
		readCounter := 0
		conn.ReadStub = func(bytes []byte) (int, error) {
			readCounter++
			if readCounter == 1 || readCounter == 2 {
				n := copy(bytes[:], "OK\n")
				return n, nil
			}
			if readCounter == 3 {
				n := copy(bytes[:], "ERR\n")
				return n, nil
			}
			return 0, errors.New("read banana")
		}
		err := reader.Read(context.Background(), make(chan []byte))
		Expect(err).NotTo(BeNil())
		Expect(conn.ReadCallCount()).To(Equal(3))
		Expect(conn.WriteCallCount()).To(Equal(4))
		Expect(string(conn.WriteArgsForCall(0))).To(Equal("636463757365723a"))
		Expect(string(conn.WriteArgsForCall(1))).To(Equal("11fb6d5a105a66c85408b8005c461d53818b736e"))
		Expect(string(conn.WriteArgsForCall(2))).To(Equal("REGISTER UUID=0f672312-e02a-11e8-8c13-cf8f48795343, TYPE=JSON"))
		Expect(string(conn.WriteArgsForCall(3))).To(Equal("REQUEST-DATA mydb.mytable"))
	})

	It("send lines to channel", func() {
		readCounter := 0
		conn.ReadStub = func(bytes []byte) (int, error) {
			readCounter++
			if readCounter <= 2 {
				n := copy(bytes[:], "OK\n")
				return n, nil
			}
			if readCounter <= 4 {
				n := copy(bytes[:], fmt.Sprintf("line %d\n", readCounter))
				return n, nil
			}
			return 0, io.EOF
		}
		ch := make(chan []byte, 10)
		err := reader.Read(context.Background(), ch)
		Expect(err).To(BeNil())
		Expect(conn.ReadCallCount()).To(Equal(5))
		Expect(conn.WriteCallCount()).To(Equal(4))
		Expect(string(conn.WriteArgsForCall(0))).To(Equal("636463757365723a"))
		Expect(string(conn.WriteArgsForCall(1))).To(Equal("11fb6d5a105a66c85408b8005c461d53818b736e"))
		Expect(string(conn.WriteArgsForCall(2))).To(Equal("REGISTER UUID=0f672312-e02a-11e8-8c13-cf8f48795343, TYPE=JSON"))
		Expect(string(conn.WriteArgsForCall(3))).To(Equal("REQUEST-DATA mydb.mytable"))
		Expect(len(ch)).To(Equal(2))
		Expect(string(<-ch)).To(Equal("line 3\n"))
		Expect(string(<-ch)).To(Equal("line 4\n"))
	})
})
