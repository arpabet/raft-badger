module go.arpabet.com/raft-badger

go 1.14

replace github.com/dgraph-io/badger => go.arpabet.com/badger/v2 v2.0.3-multi

require (
	github.com/dgraph-io/badger/v2 v2.0.3
	github.com/golang/protobuf v1.4.2
	github.com/hashicorp/raft v1.1.1
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	google.golang.org/protobuf v1.23.0
)
