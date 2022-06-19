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
	"encoding/binary"
	"github.com/dgraph-io/badger/v2"
	"google.golang.org/protobuf/proto"
	"github.com/hashicorp/raft"
)

type logStore struct {
	db          *badger.DB
	prefix      []byte
	prefixLen   int
}

func NewLogStore(db *badger.DB, prefix []byte) raft.LogStore {
	return &logStore {
		db: db,
		prefix: prefix,
		prefixLen: len(prefix),
	}
}

func (t *logStore) FirstIndex() (uint64, error) {
	first := uint64(0)
	err := t.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		it.Seek(t.prefix)
		if it.ValidForPrefix(t.prefix) {
			item := it.Item()
			rawKey := item.Key()
			if len(rawKey) > t.prefixLen {
				key := rawKey[t.prefixLen:]
				if len(key) == 8 {
					first = binary.BigEndian.Uint64(key)
				}
			}
		}
		return nil
	})
	return first, err
}

// LastIndex returns the last index written. 0 for no entries.
func (t *logStore) LastIndex() (uint64, error) {
	last := uint64(0)
	err := t.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()
		// ensure reverse seeking will include the
		// see https://github.com/dgraph-io/badger/issues/436 and
		// https://github.com/dgraph-io/badger/issues/347
		seekKey := append(t.prefix, 0xFF)
		it.Seek(seekKey)
		if it.ValidForPrefix(t.prefix) {
			item := it.Item()
			rawKey := item.Key()
			if len(rawKey) > t.prefixLen {
				key := rawKey[t.prefixLen:]
				if len(key) == 8 {
					last = binary.BigEndian.Uint64(key)
				}
			}
		}
		return nil
	})
	return last, err
}

func toRaftLog(log *raft.Log) *RaftLog {
	return &RaftLog{
		Index:                log.Index,
		Term:                 log.Term,
		Type:                 RaftLogType(int(log.Type)),
		Data:                 log.Data,
		Extensions:           log.Extensions,
	}
}

func toLog(log *RaftLog) *raft.Log {
	return &raft.Log{
		Index:                log.Index,
		Term:                 log.Term,
		Type:                 raft.LogType(int(log.Type)),
		Data:                 log.Data,
		Extensions:           log.Extensions,
	}
}

func (t *logStore) getRawKey(index uint64) []byte {
	key := make([]byte, t.prefixLen + 8)
	copy(key, t.prefix)
	binary.BigEndian.PutUint64(key[t.prefixLen:], index)
	return key
}

// GetLog gets a log entry at a given index.
func (t *logStore) GetLog(index uint64, log *raft.Log) error {
	return t.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(t.getRawKey(index))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return raft.ErrLogNotFound
			}
			return err
		}
		return item.Value(func(b []byte) error {
			var raftLog RaftLog
			if e := proto.Unmarshal(b, &raftLog); e != nil {
				return e
			}
			log.Index = raftLog.Index
			log.Term = raftLog.Term
			log.Type = raft.LogType(int(raftLog.Type))
			log.Data = raftLog.Data
			log.Extensions = raftLog.Extensions
			return nil
		})
	})
}

// StoreLog stores a log entry.
func (t *logStore) StoreLog(log *raft.Log) error {
	tnx := t.db.NewTransaction(true)
	defer tnx.Discard()

	data, err := proto.Marshal(toRaftLog(log))
	if err != nil {
		return err
	}
	if err := tnx.Set(t.getRawKey(log.Index), data); err != nil {
		return err
	}

	return tnx.Commit()
}

// StoreLogs stores multiple log entries.
func (t *logStore) StoreLogs(logs []*raft.Log) error {
	maxBatchSize := int(t.db.MaxBatchSize())

	var tnx *badger.Txn
	for i, log := range logs {
		
		if i % maxBatchSize == 0 {
			if tnx != nil {
				if err := tnx.Commit(); err != nil {
					return err
				}
				tnx = nil
			}
		}

		if tnx == nil {
			tnx = t.db.NewTransaction(true)
		}

		data, err := proto.Marshal(toRaftLog(log))
		if err != nil {
			tnx.Discard()
			return err
		}

		if err := tnx.Set(t.getRawKey(log.Index), data); err != nil {
			tnx.Discard()
			return err
		}
		
	}

	if tnx != nil {
		return tnx.Commit()
	}

	return nil
}

// DeleteRange deletes a range of log entries. The range is inclusive.
func (t *logStore) DeleteRange(min, max uint64) error {
	maxBatchSize := uint64(t.db.MaxBatchSize())

	var txn *badger.Txn
	var it  *badger.Iterator
	for index := min; index <= max; index++ {

		if index % maxBatchSize == 0 {
			if it != nil {
				it.Close()
				it = nil
			}
			if txn != nil {
				if err := txn.Commit(); err != nil {
					return err
				}
				txn = nil
			}
		}

		if txn == nil {
			txn = t.db.NewTransaction(true)
		}

		if it == nil {
			it = txn.NewIterator(badger.DefaultIteratorOptions)
			it.Seek(t.getRawKey(index))
		}

		if !it.ValidForPrefix(t.prefix) {
			break
		}

		item := it.Item()
		rawKey := item.Key()
		key := rawKey[t.prefixLen:]
		if len(key) == 8 {
			index = binary.BigEndian.Uint64(key)
			if index > max {
				break
			}

			if err := txn.Delete(rawKey); err != nil {
				it.Close()
				txn.Discard()
				return err
			}

		}

		it.Next()

	}

	if it != nil {
		it.Close()
		it = nil
	}
	if txn != nil {
		return txn.Commit()
	}
	return nil
}
