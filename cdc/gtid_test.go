// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cdc_test

import (
	"github.com/bborbe/kafka-maxscale-cdc-connector/cdc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GTID", func() {

	It("parse string", func() {
		gtid, err := cdc.ParseGTID("1-2-3")
		Expect(err).To(BeNil())
		Expect(gtid).NotTo(BeNil())
		Expect(gtid.Domain).To(Equal(uint32(1)))
		Expect(gtid.ServerId).To(Equal(uint32(2)))
		Expect(gtid.Sequence).To(Equal(uint64(3)))
	})

	It("string return ", func() {
		id := "1-2-3"
		gtid, err := cdc.ParseGTID(id)
		Expect(err).To(BeNil())
		Expect(gtid).NotTo(BeNil())
		Expect(gtid.String()).To(Equal("1-2-3"))
	})

	It("nil gtid return empty string", func() {
		var gtid *cdc.GTID
		Expect(gtid.String()).To(Equal(""))
	})

})
