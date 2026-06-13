/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package raftbadger

import (
	"encoding/binary"
	"errors"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	"google.golang.org/protobuf/proto"
)

type logStore struct {
	db        *badger.DB
	prefix    []byte
	prefixLen int
}

func NewLogStore(db *badger.DB, prefix []byte) raft.LogStore {
	return &logStore{
		db:        db,
		prefix:    prefix,
		prefixLen: len(prefix),
	}
}

// indexFromKey extracts the big-endian uint64 index from a prefixed key. It
// returns false for keys that do not carry an 8-byte index suffix.
func (t *logStore) indexFromKey(rawKey []byte) (uint64, bool) {
	if len(rawKey) != t.prefixLen+8 {
		return 0, false
	}
	return binary.BigEndian.Uint64(rawKey[t.prefixLen:]), true
}

func (t *logStore) getRawKey(index uint64) []byte {
	key := make([]byte, t.prefixLen+8)
	copy(key, t.prefix)
	binary.BigEndian.PutUint64(key[t.prefixLen:], index)
	return key
}

// FirstIndex returns the first index written. 0 for no entries.
func (t *logStore) FirstIndex() (uint64, error) {
	var first uint64
	err := t.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		it.Seek(t.prefix)
		if it.ValidForPrefix(t.prefix) {
			if idx, ok := t.indexFromKey(it.Item().Key()); ok {
				first = idx
			}
		}
		return nil
	})
	return first, err
}

// LastIndex returns the last index written. 0 for no entries.
func (t *logStore) LastIndex() (uint64, error) {
	var last uint64
	err := t.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()
		// Seek past the largest possible key for this prefix so reverse
		// iteration lands on the highest stored index. A full 8-byte 0xFF
		// suffix is required to cover indexes whose leading byte is 0xFF.
		// See https://github.com/dgraph-io/badger/issues/436
		seekKey := make([]byte, t.prefixLen+8)
		copy(seekKey, t.prefix)
		for i := t.prefixLen; i < len(seekKey); i++ {
			seekKey[i] = 0xFF
		}
		it.Seek(seekKey)
		if it.ValidForPrefix(t.prefix) {
			if idx, ok := t.indexFromKey(it.Item().Key()); ok {
				last = idx
			}
		}
		return nil
	})
	return last, err
}

func toRaftLog(log *raft.Log) *RaftLog {
	out := &RaftLog{
		Index:      log.Index,
		Term:       log.Term,
		Type:       RaftLogType(int(log.Type)),
		Data:       log.Data,
		Extensions: log.Extensions,
	}
	// Persist AppendedAt as Unix nanoseconds; leave 0 for the zero time so it
	// round-trips back to a zero time.Time rather than year 1.
	if !log.AppendedAt.IsZero() {
		out.AppendedAt = log.AppendedAt.UnixNano()
	}
	return out
}

func fromRaftLog(raftLog *RaftLog, log *raft.Log) {
	log.Index = raftLog.Index
	log.Term = raftLog.Term
	log.Type = raft.LogType(int(raftLog.Type))
	log.Data = raftLog.Data
	log.Extensions = raftLog.Extensions
	if raftLog.AppendedAt != 0 {
		log.AppendedAt = time.Unix(0, raftLog.AppendedAt).UTC()
	} else {
		log.AppendedAt = time.Time{}
	}
}

// GetLog gets a log entry at a given index.
func (t *logStore) GetLog(index uint64, log *raft.Log) error {
	return t.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(t.getRawKey(index))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return raft.ErrLogNotFound
			}
			return err
		}
		return item.Value(func(b []byte) error {
			var raftLog RaftLog
			if e := proto.Unmarshal(b, &raftLog); e != nil {
				return e
			}
			fromRaftLog(&raftLog, log)
			return nil
		})
	})
}

// StoreLog stores a log entry.
func (t *logStore) StoreLog(log *raft.Log) error {
	return t.StoreLogs([]*raft.Log{log})
}

// StoreLogs stores multiple log entries. A WriteBatch transparently flushes
// whenever badger's per-transaction entry-count or size limits are reached, so
// arbitrarily large batches are handled safely.
func (t *logStore) StoreLogs(logs []*raft.Log) error {
	wb := t.db.NewWriteBatch()
	defer wb.Cancel()

	for _, log := range logs {
		data, err := proto.Marshal(toRaftLog(log))
		if err != nil {
			return err
		}
		if err := wb.Set(t.getRawKey(log.Index), data); err != nil {
			return err
		}
	}

	return wb.Flush()
}

// DeleteRange deletes a range of log entries. The range is inclusive.
func (t *logStore) DeleteRange(min, max uint64) error {
	wb := t.db.NewWriteBatch()
	defer wb.Cancel()

	err := t.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(t.getRawKey(min)); it.ValidForPrefix(t.prefix); it.Next() {
			idx, ok := t.indexFromKey(it.Item().Key())
			if !ok {
				continue
			}
			if idx > max {
				break
			}
			// KeyCopy is required: the iterator reuses its key buffer on
			// Next(), but the WriteBatch retains the key until Flush.
			if err := wb.Delete(it.Item().KeyCopy(nil)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return wb.Flush()
}
