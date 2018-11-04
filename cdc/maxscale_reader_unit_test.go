// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cdc

import (
	"bytes"
	"testing"
)

func TestPrintAuth(t *testing.T) {
	c := &MaxscaleReader{
		User:     "cdcuser",
		Password: "cdc",
	}
	buf := &bytes.Buffer{}
	err := c.writeAuth(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "636463757365723a11fb6d5a105a66c85408b8005c461d53818b736e"
	if expected != buf.String() {
		t.Fatalf("expect %s got %s", expected, buf.String())
	}
}

func TestStartsWith(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		prefix   []byte
		expected bool
	}{
		{
			name:     "both empty",
			content:  []byte{},
			prefix:   []byte{},
			expected: true,
		},
		{
			name:     "substring",
			content:  []byte("OK 123"),
			prefix:   []byte("OK"),
			expected: true,
		},
		{
			name:     "not equal",
			content:  []byte("ERR 123"),
			prefix:   []byte("OK"),
			expected: false,
		},
		{
			name:     "no substring",
			content:  []byte("OK"),
			prefix:   []byte("OK 123"),
			expected: false,
		},
		{
			name:     "equal",
			content:  []byte("OK"),
			prefix:   []byte("OK"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := startsWith(tt.content, tt.prefix)
			if tt.expected != result {
				t.Errorf("expected %v but got %v with", tt.expected, result)
			}
		})
	}
}
