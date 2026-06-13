/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package raftbadger_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	raftbadger "go.arpabet.com/raft-badger"
)

func TestUint64Operations(t *testing.T) {
	db := openTestDB(t)
	stable := raftbadger.NewStableStore(db, []byte("conf"))

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

// TestGetUint64MaxValue guards the big-endian encoding at the boundary.
func TestGetUint64MaxValue(t *testing.T) {
	db := openTestDB(t)
	stable := raftbadger.NewStableStore(db, []byte("conf"))

	require.NoError(t, stable.SetUint64([]byte("max"), ^uint64(0)))
	val, err := stable.GetUint64([]byte("max"))
	require.NoError(t, err)
	require.Equal(t, ^uint64(0), val)
}

// TestGetUint64Corrupt verifies a non-8-byte value yields an error instead of
// panicking inside binary.BigEndian.Uint64.
func TestGetUint64Corrupt(t *testing.T) {
	db := openTestDB(t)
	stable := raftbadger.NewStableStore(db, []byte("conf"))

	require.NoError(t, stable.Set([]byte("short"), []byte("xyz")))
	_, err := stable.GetUint64([]byte("short"))
	require.Error(t, err)

	// An explicitly empty value is treated as the zero value, not corruption.
	require.NoError(t, stable.Set([]byte("blank"), []byte{}))
	val, err := stable.GetUint64([]byte("blank"))
	require.NoError(t, err)
	require.Equal(t, uint64(0), val)
}

// TestStableStorePrefixIsolation ensures distinct prefixes do not collide,
// including the README pattern of a log store and a stable store sharing a DB.
func TestStableStorePrefixIsolation(t *testing.T) {
	db := openTestDB(t)
	a := raftbadger.NewStableStore(db, []byte("a"))
	b := raftbadger.NewStableStore(db, []byte("b"))

	require.NoError(t, a.Set([]byte("k"), []byte("av")))
	require.NoError(t, b.Set([]byte("k"), []byte("bv")))

	va, err := a.Get([]byte("k"))
	require.NoError(t, err)
	require.Equal(t, []byte("av"), va)

	vb, err := b.Get([]byte("k"))
	require.NoError(t, err)
	require.Equal(t, []byte("bv"), vb)
}

// TestStableStorePrefixBackingNotMutated is a regression test for an aliasing
// bug: Set/getImpl used append(t.prefix, key...), which writes into the
// caller's prefix backing array when it has spare capacity, corrupting it and
// racing under concurrency. Sentinel bytes in the spare capacity must survive.
func TestStableStorePrefixBackingNotMutated(t *testing.T) {
	db := openTestDB(t)

	backing := make([]byte, 4, 32)
	copy(backing, "conf")
	full := backing[:cap(backing)]
	for i := 4; i < len(full); i++ {
		full[i] = 'G'
	}
	prefix := backing[:4]

	stable := raftbadger.NewStableStore(db, prefix)
	require.NoError(t, stable.Set([]byte("longerkey"), []byte("v")))
	require.NoError(t, stable.SetUint64([]byte("num"), 42))
	_, err := stable.Get([]byte("longerkey"))
	require.NoError(t, err)
	_, err = stable.GetUint64([]byte("num"))
	require.NoError(t, err)

	require.Equal(t, "conf", string(prefix), "prefix contents changed")
	for i := 4; i < len(full); i++ {
		require.Equalf(t, byte('G'), full[i], "prefix backing array mutated at offset %d", i)
	}
}

// TestStableStoreConcurrent exercises Set/Get from multiple goroutines; run
// with -race it guards against the prefix-aliasing data race.
func TestStableStoreConcurrent(t *testing.T) {
	db := openTestDB(t)
	stable := raftbadger.NewStableStore(db, []byte("conf"))

	const writers = 8
	const perWriter = 200

	errCh := make(chan error, writers)
	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < perWriter; i++ {
				key := []byte(fmt.Sprintf("w%d-k%d", w, i))
				want := []byte(fmt.Sprintf("w%d-v%d", w, i))
				if err := stable.Set(key, want); err != nil {
					errCh <- err
					return
				}
				got, err := stable.Get(key)
				if err != nil {
					errCh <- err
					return
				}
				if string(got) != string(want) {
					errCh <- fmt.Errorf("key %q: got %q want %q", key, got, want)
					return
				}
			}
		}(w)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	// Every key written must still be readable with its expected value.
	for w := 0; w < writers; w++ {
		for i := 0; i < perWriter; i++ {
			got, err := stable.Get([]byte(fmt.Sprintf("w%d-k%d", w, i)))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprintf("w%d-v%d", w, i), string(got))
		}
	}
}
