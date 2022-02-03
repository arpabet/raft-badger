/*
 *
 * Copyright 2020-present Arpabet, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package raft_badger

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func TestLogOperations(t *testing.T) {

	fd, err := ioutil.TempFile(os.TempDir(), "raftbadger-test")
	require.NoError(t, err)
	filePath := fd.Name()
	fd.Close()
	os.Remove(filePath)

	db, err := badger.Open(badger.DefaultOptions(filePath))
	require.NoError(t, err)

	defer func() {
		db.Close()
		os.RemoveAll(filePath)
	}()

	log := NewLogStore(db, []byte("log"))

	first, err := log.FirstIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(0), first)

	last, err := log.LastIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(0), last)

	var entry raft.Log
	err = log.GetLog(uint64(100), &entry)
	require.Equal(t, raft.ErrLogNotFound, err)

	entry.Index = 123
	entry.Data = []byte("alex")
	err = log.StoreLog(&entry)
	require.NoError(t, err)

	first, err = log.FirstIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(123), first)

	last, err = log.LastIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(123), last)

	entry.Index = 124
	entry.Data = []byte("lex")
	err = log.StoreLogs([]*raft.Log{ &entry })
	require.NoError(t, err)

	first, err = log.FirstIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(123), first)

	last, err = log.LastIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(124), last)

	err = log.DeleteRange(uint64(0), uint64(123))
	require.NoError(t, err)

	first, err = log.FirstIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(124), first)

	last, err = log.LastIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(124), last)

}
