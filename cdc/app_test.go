// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cdc_test

import (
	"github.com/bborbe/kafka-maxscale-cdc-connector/cdc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CDC App", func() {
	var app *cdc.App
	BeforeEach(func() {
		app = &cdc.App{
			CdcDatabase:  "mydb",
			CdcFormat:    "JSON",
			CdcHost:      "myhost",
			CdcPassword:  "mypass",
			CdcPort:      4001,
			CdcTable:     "mytable",
			CdcUser:      "myuser",
			CdcUUID:      "0f672312-e02a-11e8-8c13-cf8f48795343",
			KafkaBrokers: "kafka:9092",
			KafkaTopic:   "mytopic",
			Port:         8080,
		}
	})
	It("Validate without error", func() {
		Expect(app.Validate()).NotTo(HaveOccurred())
	})
	It("Validate returns error if CdcDatabase is empty", func() {
		app.CdcDatabase = ""
		Expect(app.Validate()).To(HaveOccurred())
	})
	It("Validate returns error if CdcFormat not JSON or AVRO", func() {
		app.CdcFormat = "banana"
		Expect(app.Validate()).To(HaveOccurred())
	})
	It("Validate returns no error if CdcFormat is JSON", func() {
		app.CdcFormat = "JSON"
		Expect(app.Validate()).NotTo(HaveOccurred())
	})
	It("Validate returns no error if CdcFormat is AVRO", func() {
		app.CdcFormat = "AVRO"
		Expect(app.Validate()).NotTo(HaveOccurred())
	})
	It("Validate returns error if CdcHost is empty", func() {
		app.CdcHost = ""
		Expect(app.Validate()).To(HaveOccurred())
	})
	It("Validate returns error if CdcPassword is empty", func() {
		app.CdcPassword = ""
		Expect(app.Validate()).To(HaveOccurred())
	})
	It("Validate returns error if CdcPort is 0", func() {
		app.CdcPort = 0
		Expect(app.Validate()).To(HaveOccurred())
	})
	It("Validate returns error if CdcTable is empty", func() {
		app.CdcTable = ""
		Expect(app.Validate()).To(HaveOccurred())
	})
	It("Validate returns error if CdcUser is empty", func() {
		app.CdcUser = ""
		Expect(app.Validate()).To(HaveOccurred())
	})
	It("Validate returns error if CdcUUID is empty", func() {
		app.CdcUUID = ""
		Expect(app.Validate()).To(HaveOccurred())
	})
	It("Validate returns error if KafkaBrokers is empty", func() {
		app.KafkaBrokers = ""
		Expect(app.Validate()).To(HaveOccurred())
	})
	It("Validate returns error if KafkaTopic is empty", func() {
		app.KafkaTopic = ""
		Expect(app.Validate()).To(HaveOccurred())
	})
	It("Validate returns error if Port is 0", func() {
		app.Port = 0
		Expect(app.Validate()).To(HaveOccurred())
	})

})
