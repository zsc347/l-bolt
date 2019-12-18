package bolt

import "fmt"

import "unsafe"

// DB represents a collection of buckets persisted to a file on disk.
// All data access is performed through transactions which can be obtained
// through the DB.
// All the functions on DB will return a ErrDatabaseNotOpen if accessed
// before Open is called.
type DB struct {
	pageSize int
	data     *[maxMapSize]byte
}

// page retrieves a page reference from the mmap based on the current page size
func (db *DB) page(id pgid) *page {
	pos := id * pgid(db.pageSize)
	return (*page)(unsafe.Pointer(&db.data[pos]))
}

type meta struct {
	magic    uint32
	version  uint32
	pageSize uint32
	flags    uint32
	root     bucket
	freelist pgid
	pgid     pgid
	txid     txid
	checksum uint64
}

// _assert will panic with a given formatted message if the given condition is false
func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}
