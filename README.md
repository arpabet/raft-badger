# raft-badger

A [BadgerDB](https://github.com/dgraph-io/badger) backend for [HashiCorp Raft](https://github.com/hashicorp/raft).

`raft-badger` implements the `raft.LogStore` and `raft.StableStore` interfaces on
top of a single BadgerDB instance, so Raft's replicated log and its stable
configuration can be persisted in one embedded key-value store.

## Features

- Implements `raft.LogStore` (log entries) and `raft.StableStore` (key/value and
  `uint64` config).
- Both stores can share **one** `*badger.DB`; each is namespaced by a key
  prefix, so they never collide.
- Log entries are protobuf-encoded and preserve every `raft.Log` field,
  including `Type`, `Extensions`, and `AppendedAt`.
- Safe for concurrent use.
- Bulk writes (`StoreLogs`, `DeleteRange`) use Badger `WriteBatch`, which flushes
  automatically when per-transaction limits are reached.

## Requirements

- Go **1.25+**
- [`github.com/dgraph-io/badger/v4`](https://github.com/dgraph-io/badger)
- [`github.com/hashicorp/raft`](https://github.com/hashicorp/raft) v1.7+

## Install

```bash
go get go.arpabet.com/raft-badger
```

## Usage

Open a BadgerDB instance and back both Raft stores with it, giving each a
distinct prefix:

```go
import (
	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	raftbadger "go.arpabet.com/raft-badger"
)

// A single database backs both stores.
db, err := badger.Open(badger.DefaultOptions("/var/lib/myapp/raft"))
if err != nil {
	return err
}
defer db.Close()

// Use distinct, non-overlapping prefixes for the two stores.
logStore := raftbadger.NewLogStore(db, []byte("log"))
stableStore := raftbadger.NewStableStore(db, []byte("conf"))

// snapshots, transport and the FSM are provided by your application
// and hashicorp/raft (raft.NewFileSnapshotStore, raft.NewTCPTransport, ...).
r, err := raft.NewRaft(raft.DefaultConfig(), fsm, logStore, stableStore, snapshots, transport)
if err != nil {
	return err
}
```

## API

```go
// NewLogStore returns a raft.LogStore whose keys are namespaced by prefix.
func NewLogStore(db *badger.DB, prefix []byte) raft.LogStore

// NewStableStore returns a raft.StableStore whose keys are namespaced by prefix.
func NewStableStore(db *badger.DB, prefix []byte) raft.StableStore
```

## Storage layout

- **Log entries** are stored under `prefix + <8-byte big-endian index>`, with the
  value being a protobuf-encoded `RaftLog`.
- **Stable values** are stored under `prefix + key`.

When sharing one database, choose prefixes that are not a prefix of one another
(e.g. `"log"` and `"conf"`) so iteration over one store never sees the other's
keys.

## Development

```bash
make test    # go test -race -cover ./...
make build   # test, then go build ./...
```

Regenerating the protobuf code (only needed when editing `raftbadger.proto`)
requires `protoc` and `protoc-gen-go` on your `PATH`:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
make proto
```

## License

Business Source License 1.1 (BUSL-1.1). The Licensed Work is © 2025 Karagatan
LLC; the Change License is MPL 2.0. See [LICENSE](LICENSE) for the full terms.
