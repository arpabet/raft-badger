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
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
)

type stableStore struct {
	db          *badger.DB
	prefix      []byte
}

func NewStableStore(db *badger.DB, prefix []byte) raft.StableStore {
	return &stableStore {
		db: db,
		prefix: prefix,
	}
}

func (t* stableStore) Set(key []byte, val []byte) error {

	tx := t.db.NewTransaction(true)
	defer tx.Discard()

	entry := &badger.Entry{ Key: append(t.prefix, key...), Value: val }

	err := tx.SetEntry(entry)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Get returns the value for key, or an empty byte slice if key was not found.
func (t* stableStore) Get(key []byte) ([]byte, error) {
	return t.getImpl(key, nil)
}

func (t* stableStore) getImpl(key, dst []byte) ([]byte, error) {

	tx := t.db.NewTransaction(false)
	defer tx.Discard()

	rawKey := append(t.prefix, key...)

	item, err := tx.Get(rawKey)
	if err != nil {

		if err == badger.ErrKeyNotFound {
			return dst, nil
		}

		return nil, err
	}

	data, err := item.ValueCopy(dst)
	if err != nil {
		return nil, errors.Errorf("badger fetch value failed '%v', %v", rawKey, err)
	}

	return data, nil

}

func (t* stableStore) SetUint64(key []byte, val uint64) error {

	var data [8]byte
	buf := data[:]

	binary.BigEndian.PutUint64(buf, val)

	return t.Set(key, buf)

}

// GetUint64 returns the uint64 value for key, or 0 if key was not found.
func (t* stableStore) GetUint64(key []byte) (uint64, error) {

	var data [8]byte

	buf, err := t.getImpl(key, data[:])
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint64(buf), nil
}
