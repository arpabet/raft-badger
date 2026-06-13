/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package raftbadger_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/require"
	raftbadger "go.arpabet.com/raft-badger"
)

func TestLogOperations(t *testing.T) {
	db := openTestDB(t)
	log := raftbadger.NewLogStore(db, []byte("log"))

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
	err = log.StoreLogs([]*raft.Log{&entry})
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

// TestGetLogAllFields verifies every persisted field of raft.Log round-trips.
func TestGetLogAllFields(t *testing.T) {
	db := openTestDB(t)
	log := raftbadger.NewLogStore(db, []byte("log"))

	in := &raft.Log{
		Index:      7,
		Term:       3,
		Type:       raft.LogConfiguration,
		Data:       []byte("payload"),
		Extensions: []byte("extension-bytes"),
	}
	require.NoError(t, log.StoreLog(in))

	var out raft.Log
	require.NoError(t, log.GetLog(7, &out))
	require.Equal(t, in.Index, out.Index)
	require.Equal(t, in.Term, out.Term)
	require.Equal(t, in.Type, out.Type)
	require.Equal(t, in.Data, out.Data)
	require.Equal(t, in.Extensions, out.Extensions)
	// An unset AppendedAt must round-trip back to the zero time, not year 1.
	require.True(t, out.AppendedAt.IsZero())
}

// TestAppendedAtRoundTrip verifies the AppendedAt timestamp is persisted and
// restored. Previously it was silently dropped (absent from the proto schema).
func TestAppendedAtRoundTrip(t *testing.T) {
	db := openTestDB(t)
	log := raftbadger.NewLogStore(db, []byte("log"))

	appendedAt := time.Unix(1700000000, 123456789).UTC()
	require.NoError(t, log.StoreLog(&raft.Log{Index: 1, Term: 2, AppendedAt: appendedAt}))

	var out raft.Log
	require.NoError(t, log.GetLog(1, &out))
	require.False(t, out.AppendedAt.IsZero())
	require.True(t, appendedAt.Equal(out.AppendedAt), "want %s, got %s", appendedAt, out.AppendedAt)
	require.Equal(t, appendedAt.UnixNano(), out.AppendedAt.UnixNano())
}

// TestStoreLogsLargeBatch stores more entries than fit in a single badger
// transaction, exercising the WriteBatch auto-flush path. The previous
// implementation used db.MaxBatchSize() (a byte budget) as an entry-count
// modulus, which funneled every entry into one transaction and failed with
// ErrTxnTooBig once the batch exceeded the per-transaction entry limit.
func TestStoreLogsLargeBatch(t *testing.T) {
	db := openTestDB(t)
	log := raftbadger.NewLogStore(db, []byte("log"))

	n := int(db.MaxBatchCount()) + 1000
	logs := make([]*raft.Log, n)
	for i := range logs {
		logs[i] = &raft.Log{
			Index: uint64(i + 1),
			Term:  1,
			Data:  []byte(fmt.Sprintf("entry-%d", i+1)),
		}
	}
	require.NoError(t, log.StoreLogs(logs))

	first, err := log.FirstIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(1), first)

	last, err := log.LastIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(n), last)

	// Spot-check the first, a middle, and the last entry.
	for _, idx := range []uint64{1, uint64(n / 2), uint64(n)} {
		var out raft.Log
		require.NoError(t, log.GetLog(idx, &out))
		require.Equal(t, idx, out.Index)
		require.Equal(t, []byte(fmt.Sprintf("entry-%d", idx)), out.Data)
	}
}

func TestDeleteRange(t *testing.T) {
	storeLogs := func(t *testing.T, log raft.LogStore, from, to uint64) {
		logs := make([]*raft.Log, 0, to-from+1)
		for i := from; i <= to; i++ {
			logs = append(logs, &raft.Log{Index: i, Term: 1, Data: []byte(fmt.Sprintf("e-%d", i))})
		}
		require.NoError(t, log.StoreLogs(logs))
	}

	assertBounds := func(t *testing.T, log raft.LogStore, wantFirst, wantLast uint64) {
		first, err := log.FirstIndex()
		require.NoError(t, err)
		require.Equal(t, wantFirst, first)
		last, err := log.LastIndex()
		require.NoError(t, err)
		require.Equal(t, wantLast, last)
	}

	assertMissing := func(t *testing.T, log raft.LogStore, indexes ...uint64) {
		for _, idx := range indexes {
			var out raft.Log
			require.Equal(t, raft.ErrLogNotFound, log.GetLog(idx, &out), "index %d should be deleted", idx)
		}
	}

	assertPresent := func(t *testing.T, log raft.LogStore, indexes ...uint64) {
		for _, idx := range indexes {
			var out raft.Log
			require.NoError(t, log.GetLog(idx, &out), "index %d should be present", idx)
		}
	}

	t.Run("middle", func(t *testing.T) {
		log := raftbadger.NewLogStore(openTestDB(t), []byte("log"))
		storeLogs(t, log, 1, 10)
		require.NoError(t, log.DeleteRange(3, 5))
		assertMissing(t, log, 3, 4, 5)
		assertPresent(t, log, 1, 2, 6, 7, 8, 9, 10)
		assertBounds(t, log, 1, 10)
	})

	t.Run("prefix", func(t *testing.T) {
		log := raftbadger.NewLogStore(openTestDB(t), []byte("log"))
		storeLogs(t, log, 1, 10)
		require.NoError(t, log.DeleteRange(1, 4))
		assertMissing(t, log, 1, 2, 3, 4)
		assertBounds(t, log, 5, 10)
	})

	t.Run("suffix past max", func(t *testing.T) {
		log := raftbadger.NewLogStore(openTestDB(t), []byte("log"))
		storeLogs(t, log, 1, 10)
		require.NoError(t, log.DeleteRange(8, 1000))
		assertMissing(t, log, 8, 9, 10)
		assertBounds(t, log, 1, 7)
	})

	t.Run("all", func(t *testing.T) {
		log := raftbadger.NewLogStore(openTestDB(t), []byte("log"))
		storeLogs(t, log, 1, 10)
		require.NoError(t, log.DeleteRange(0, 1000))
		assertBounds(t, log, 0, 0)
	})

	t.Run("empty range is a no-op", func(t *testing.T) {
		log := raftbadger.NewLogStore(openTestDB(t), []byte("log"))
		storeLogs(t, log, 1, 10)
		require.NoError(t, log.DeleteRange(50, 60))
		assertPresent(t, log, 1, 5, 10)
		assertBounds(t, log, 1, 10)
	})
}

// TestLogStorePrefixIsolation ensures two stores sharing one database but using
// different prefixes never observe each other's keys.
func TestLogStorePrefixIsolation(t *testing.T) {
	db := openTestDB(t)
	a := raftbadger.NewLogStore(db, []byte("a"))
	b := raftbadger.NewLogStore(db, []byte("b"))

	require.NoError(t, a.StoreLog(&raft.Log{Index: 1, Data: []byte("a1")}))
	require.NoError(t, b.StoreLog(&raft.Log{Index: 2, Data: []byte("b2")}))

	first, _ := a.FirstIndex()
	last, _ := a.LastIndex()
	require.Equal(t, uint64(1), first)
	require.Equal(t, uint64(1), last)

	first, _ = b.FirstIndex()
	last, _ = b.LastIndex()
	require.Equal(t, uint64(2), first)
	require.Equal(t, uint64(2), last)

	var out raft.Log
	require.Equal(t, raft.ErrLogNotFound, a.GetLog(2, &out))
	require.Equal(t, raft.ErrLogNotFound, b.GetLog(1, &out))
}

// TestLogStorePrefixBackingNotMutated is a regression test for an aliasing bug:
// LastIndex used append(t.prefix, 0xFF), which writes into the caller's prefix
// backing array whenever it has spare capacity. The sentinel bytes living in
// that spare capacity must remain untouched.
func TestLogStorePrefixBackingNotMutated(t *testing.T) {
	db := openTestDB(t)

	backing := make([]byte, 3, 32)
	copy(backing, "log")
	full := backing[:cap(backing)]
	for i := 3; i < len(full); i++ {
		full[i] = 'G'
	}
	prefix := backing[:3]

	log := raftbadger.NewLogStore(db, prefix)
	require.NoError(t, log.StoreLog(&raft.Log{Index: 1, Data: []byte("x")}))

	_, err := log.LastIndex()
	require.NoError(t, err)
	_, err = log.FirstIndex()
	require.NoError(t, err)
	require.NoError(t, log.DeleteRange(1, 1))

	require.Equal(t, "log", string(prefix), "prefix contents changed")
	for i := 3; i < len(full); i++ {
		require.Equalf(t, byte('G'), full[i], "prefix backing array mutated at offset %d", i)
	}
}

// TestLogStoreConcurrent exercises the stores from multiple goroutines; run
// with -race it guards against the prefix-aliasing data race.
func TestLogStoreConcurrent(t *testing.T) {
	db := openTestDB(t)
	log := raftbadger.NewLogStore(db, []byte("log"))

	const writers = 8
	const perWriter = 200

	errCh := make(chan error, writers)
	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < perWriter; i++ {
				idx := uint64(w*perWriter + i + 1)
				if err := log.StoreLog(&raft.Log{Index: idx, Term: 1, Data: []byte(fmt.Sprintf("e-%d", idx))}); err != nil {
					errCh <- err
					return
				}
				if _, err := log.LastIndex(); err != nil {
					errCh <- err
					return
				}
				if _, err := log.FirstIndex(); err != nil {
					errCh <- err
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

	last, err := log.LastIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(writers*perWriter), last)

	for idx := uint64(1); idx <= uint64(writers*perWriter); idx++ {
		var out raft.Log
		require.NoError(t, log.GetLog(idx, &out))
		require.Equal(t, idx, out.Index)
	}
}
