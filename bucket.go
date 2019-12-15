package bolt

import "unsafe"

const (
	// MaxKeySize is the maximum length of a key, in bytes.
	MaxKeySize = 32768 // 1 << 15

	// MaxValueSize is the maximum length of a value, in bytes.
	MaxValueSize = (1 << 31) - 2
)

const (
	maxUint = ^uint(0)
	minUint = 0
	maxInt  = int(^uint(0) >> 1)
	minInt  = -maxInt - 1
)

const bucketHeaderSize = int(unsafe.Sizeof(bucket{}))

const (
	minFillPercent = 0.1
	maxFillPercent = 1.0
)

// DefaultFillPercent is the percentage that split pages are filled
// This value can be changed by setting Bucker.FillPercent.
const DefaultFillPercent = 0.5

// Bucket represents a collection of key/value pairs inside the database.
type Bucket struct {
	*bucket
	tx          *Tx                // the associated transaction
	buckets     map[string]*Bucket // subbucker cache
	page        *page              // inline page reference
	rootNode    *node              // materialized note for the root page
	nodes       map[pgid]*node     // node cache
	FillPercent float64
}

type bucket struct {
	root     pgid
	sequence uint64
}

// newBucket returns a new bucker associated with a transaction.
func newBucket(tx *Tx) Bucket {
	var b = Bucket{tx: tx, FillPercent: DefaultFillPercent}
	if tx.writable {
		b.buckets = make(map[string]*Bucket)
		b.nodes = make(map[pgid]*node)
	}
	return b
}

// Tx returns the tx of the bucket.
func (b *Bucket) Tx() *Tx {
	return b.tx
}

// Root returns the root of the bucket.
func (b *Bucket) Root() uint64 {
	return uint64(b.root)
}

// Writable returns whether the bucket is writable
func (b *Bucket) Writable() bool {
	return b.tx.writable
}

