/**
    Copyright (c) 2020-2022 Arpabet, Inc.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in
	all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
	THE SOFTWARE.
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
