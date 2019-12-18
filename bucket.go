package bolt

import "unsafe"

import "fmt"

import "bytes"

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

// Cursor creates a cursor associated with the bucket.
// The cursor is only valid as long as the transaction is open.
// Do not use a cursor after the transaction is closed.
func (b *Bucket) Cursor() *Cursor {
	b.tx.stats.CursorCount++

	// Allocate and return a cursor
	return &Cursor{
		bucket: b,
		stack:  make([]elemRef, 0),
	}
}

// pageNode returns the in-memory node, if it exists.
// Otherwise returns the underlying page.
func (b *Bucket) pageNode(id pgid) (*page, *node) {
	// Inline buckets have a fake page embedded in their value so treat them
	// differently.
	// We'll return the rootNode (if avaliable) or the fake page.
	if b.root == 0 {
		if id != 0 {
			panic(fmt.Sprintf("inline bucket non-zero page access(2): %d != 0", id))
		}

		if b.rootNode != nil {
			return nil, b.rootNode
		}
		return b.page, nil
	}

	if b.nodes != nil {
		if n := b.nodes[id]; n != nil {
			return nil, n
		}
	}

	// Finally lookup the page from the transaction if no node is materialized
	return b.tx.page(id), nil
}

// Bucket retrieves a nested bucket by name.
// Returns nil if the bucket does not exist.
// The bucket instance is only valid for the lifetime of the transaction
func (b *Bucket) Bucket(name []byte) *Bucket {
	if b.buckets != nil {
		if child := b.buckets[string(name)]; child != nil {
			return child
		}
	}

	// Move cursor to key
	c := b.Cursor()
	k, v, flags := c.seek(name)

	// Return nil if the key donesn't exist or it is not a bucket
	if !bytes.Equal(name, k) || (flags&bucketLeafFlag) == 0 {
		return nil
	}

	var child = b.openBucket(v)
	// Otherwise create a bucket and cache it.
	if b.buckets != nil {
		b.buckets[string(name)] = child
	}
	return child
}

// cloneBytes returns a copy of a given slice
func cloneBytes(v []byte) []byte {
	var clone = make([]byte, len(v))
	copy(clone, v)
	return clone
}

// Helper method that re-interprets a sub-bucket value
// from a parent into a Bucket
func (b *Bucket) openBucket(value []byte) *Bucket {
	var child = newBucket(b.tx)

	// If unaligned load/stores are broken on this arch and value is
	// unaligned simply clone to an aligned byte array
	unaligned := brokenUnaligned && uintptr(unsafe.Pointer(&value[0]))&3 != 0

	if unaligned {
		value = cloneBytes(value)
	}

	// If this is a writable transaction, then we need to copy the bucket entry.
	// Read-only transactions can point directly at the mmap entry
	if b.tx.writable && !unaligned {
		child.bucket = &bucket{}
		*child.bucket = *(*bucket)(unsafe.Pointer(&value[0]))
	} else {
		child.bucket = (*bucket)(unsafe.Pointer(&value[0]))
	}

	// Save a reference to the inline page if the bucket is inline
	if child.root == 0 {
		child.page = (*page)(unsafe.Pointer(&value[bucketHeaderSize]))
	}

	return &child
}
