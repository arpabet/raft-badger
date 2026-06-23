/*
 * Copyright (c) 2025 Karagatan LLC.
 * SPDX-License-Identifier: BUSL-1.1
 */

package raftbadger

import (
	"encoding/binary"
	"errors"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	"golang.org/x/xerrors"
)

type stableStore struct {
	db        *badger.DB
	prefix    []byte
	prefixLen int
}

func NewStableStore(db *badger.DB, prefix []byte) raft.StableStore {
	return &stableStore{
		db:        db,
		prefix:    prefix,
		prefixLen: len(prefix),
	}
}

// rawKey returns a freshly allocated prefix+key buffer. It must never alias the
// shared prefix slice; doing so (e.g. via append(t.prefix, key...)) corrupts the
// backing array and races when the store is used from multiple goroutines.
func (t *stableStore) rawKey(key []byte) []byte {
	out := make([]byte, t.prefixLen+len(key))
	copy(out, t.prefix)
	copy(out[t.prefixLen:], key)
	return out
}

func (t *stableStore) Set(key []byte, val []byte) error {
	tx := t.db.NewTransaction(true)
	defer tx.Discard()

	if err := tx.SetEntry(badger.NewEntry(t.rawKey(key), val)); err != nil {
		return err
	}

	return tx.Commit()
}

// Get returns the value for key, or an empty byte slice if key was not found.
func (t *stableStore) Get(key []byte) ([]byte, error) {
	return t.getImpl(key, nil)
}

func (t *stableStore) getImpl(key, dst []byte) ([]byte, error) {
	tx := t.db.NewTransaction(false)
	defer tx.Discard()

	rawKey := t.rawKey(key)

	item, err := tx.Get(rawKey)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return dst, nil
		}
		return nil, err
	}

	data, err := item.ValueCopy(dst)
	if err != nil {
		return nil, xerrors.Errorf("raftbadger: fetch value for key %q failed: %w", rawKey, err)
	}

	// Coalesce an empty stored value with the missing-key case so callers see a
	// single, predictable "no value" result (nil for Get) regardless of the
	// underlying engine's empty-slice representation.
	if len(data) == 0 {
		return dst, nil
	}

	return data, nil
}

func (t *stableStore) SetUint64(key []byte, val uint64) error {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], val)
	return t.Set(key, buf[:])
}

// GetUint64 returns the uint64 value for key, or 0 if key was not found.
func (t *stableStore) GetUint64(key []byte) (uint64, error) {
	var buf [8]byte

	data, err := t.getImpl(key, buf[:0])
	if err != nil {
		return 0, err
	}

	// Missing keys decode as 0, mirroring the raft.StableStore contract.
	if len(data) == 0 {
		return 0, nil
	}
	if len(data) != 8 {
		return 0, xerrors.Errorf("raftbadger: corrupt uint64 value for key %q: have %d bytes, want 8", key, len(data))
	}

	return binary.BigEndian.Uint64(data), nil
}
