package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Maxscale CDC Connector", func() {
	It("Compiles", func() {
		var err error
		_, err = gexec.Build("github.com/bborbe/maxscale-cdc-connector")
		Expect(err).NotTo(HaveOccurred())
	})
})

func TestMaxscaleCDCConnector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Maxscale CDC Connector Suite")
}
