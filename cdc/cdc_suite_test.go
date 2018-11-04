package cdc_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCdc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cdc Suite")
}
