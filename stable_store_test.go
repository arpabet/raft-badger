/**
  Copyright (c) 2022 Arpabet, LLC. All rights reserved.
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


