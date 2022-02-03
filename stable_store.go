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
