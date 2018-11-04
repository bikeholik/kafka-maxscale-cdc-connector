package cdc

import (
	"bytes"
	"testing"
)

func TestPrintAuth(t *testing.T) {
	c := &Reader{
		User:     "cdcuser",
		Password: "cdc",
	}
	buf := &bytes.Buffer{}
	c.writeAuth(buf)
	expected := "636463757365723a11fb6d5a105a66c85408b8005c461d53818b736e"
	if expected != buf.String() {
		t.Fatalf("expect %s got %s", expected, buf.String())
	}
}
