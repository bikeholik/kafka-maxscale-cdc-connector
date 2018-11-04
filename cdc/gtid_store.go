// Copyright (c) 2018 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cdc

import (
	"io/ioutil"
	"path"

	"github.com/pkg/errors"
)

// GTIDStore save GTID to disk
type GTIDStore struct {
	DataDir string
}

// Read the last GTID from disk
func (g *GTIDStore) Read() (*GTID, error) {
	content, err := ioutil.ReadFile(g.path())
	if err != nil {
		return nil, errors.Wrapf(err, "read file %s failed", g.path())
	}
	return ParseGTID(string(content))
}

// Write the given GTID to disk
func (g *GTIDStore) Write(gtid *GTID) error {
	return ioutil.WriteFile(g.path(), []byte(gtid.String()), 0600)
}

func (g *GTIDStore) path() string {
	return path.Join(g.DataDir, "lastgtid")
}
