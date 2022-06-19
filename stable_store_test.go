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
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func TestUint64Operations(t *testing.T) {

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

	stable := NewStableStore(db, []byte("conf"))

	val, err := stable.GetUint64([]byte("empty"))
	require.NoError(t, err)
	require.Equal(t, uint64(0), val)

	err = stable.SetUint64([]byte("one"), uint64(1))
	require.NoError(t, err)

	val, err = stable.GetUint64([]byte("one"))
	require.NoError(t, err)
	require.Equal(t, uint64(1), val)

	err = stable.Set([]byte("two"), []byte("val"))
	require.NoError(t, err)

	v, err := stable.Get([]byte("two"))
	require.NoError(t, err)
	require.Equal(t, []byte("val"), v)

	err = stable.Set([]byte("two"), nil)
	require.NoError(t, err)

	v, err = stable.Get([]byte("two"))
	require.NoError(t, err)
	require.Nil(t, v)

	v, err = stable.Get([]byte("three"))
	require.NoError(t, err)
	require.Nil(t, v)

	err = stable.Set([]byte("five"), []byte("value"))
	require.NoError(t, err)

}


