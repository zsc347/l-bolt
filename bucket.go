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

// node creates a node from a page and associates it with a given parent.
func (b *Bucket) node(pgid pgid, parent *node) *node {
	// TODO
	return nil
}

// write allocates and writes a bucket to a byte slice
func (b *Bucket) write() []byte {
	// Allocate the appropriate size.
	var n = b.rootNode
	var value = make([]byte, bucketHeaderSize+n.size())

	// Write a bucket header.
	var bucket = (*bucket)(unsafe.Pointer(&value[0]))
	*bucket = *b.bucket

	// Convert byte slice to a fake page and write the root node.
	var p = (*page)(unsafe.Pointer(&value[bucketHeaderSize]))
	n.write(p)

	return value
}

// CreateBucket creates a new bucket at the given key and returns the new bucket.
// Returns an error if the key already exists, if the bucker name is blank,
// or if the bucket name is too long.
// The bucket instance is only valid for the lifetime of the transaction.
func (b *Bucket) CreateBucket(key []byte) (*Bucket, error) {
	if b.tx.db == nil {
		return nil, ErrTxClosed
	} else if !b.tx.writable {
		return nil, ErrTxNotWritable
	} else if len(key) == 0 {
		return nil, ErrBucketNameRequied
	}

	// Move cursor to correct position.
	c := b.Cursor()
	k, _, flags := c.seek(key)

	// Return an error if there is an existing key.
	if bytes.Equal(key, k) {
		if (flags & bucketLeafFlag) != 0 {
			return nil, ErrBucketExists
		}
		return nil, ErrIncompatibleValue
	}

	// Create empty, inline bucket.
	var bucket = Bucket{
		bucket:      &bucket{},
		rootNode:    &node{isLeaf: true},
		FillPercent: DefaultFillPercent,
	}
	var value = bucket.write()

	// Insert into node.
	key = cloneBytes(key)
	c.node().put(key, key, value, 0, bucketLeafFlag)

	// Since sub buckets are not allowed on inline buckets, we need to
	// dereference the inline page, if it exists.
	// This will cause the bucket to be treated as a regular,
	// non-inline bucket bucket for the rest of the tx.
	b.page = nil

	return b.Bucket(key), nil
}

// CreateBucketIfNotExists creates a new bucket if it doesn't already
// exist and returns a reference to it.
// Returns an error if the bucket name is blank, or if the bucket name is too long.
// The bucket instance is only valid for the lifetime of the transaction.
func (b *Bucket) CreateBucketIfNotExists(key []byte) (*Bucket, error) {
	child, err := b.CreateBucket(key)
	if err == ErrBucketExists {
		return b.Bucket(key), nil
	} else if err != nil {
		return nil, err
	}
	return child, nil
}

// DeleteBucket deletes a bucket at a given key.
// Returns an error if the bucket does not exists,
// or if the key represents a non-bucket value.
func (b *Bucket) DeleteBucket(key []byte) error {
	if b.tx.db == nil {
		return ErrTxClosed
	} else if !b.Writable() {
		return ErrTxNotWritable
	}

	// Move Cursor to correct position.
	c := b.Cursor()
	k, _, flags := c.seek(key)

	// Return an error if bucket doesn't exist or is not a bucket
	if !bytes.Equal(key, k) {
		return ErrBucketNotFound
	} else if (flags & bucketLeafFlag) == 0 {
		return ErrIncompatibleValue
	}

	// TODO

	return nil
}

// ForEach executes a function for each key/value pair in a bucket.
// If the provided function returns an error then the iteration is stopped
// and the error is returned to the caller.
// The provided function must not modify the bucket
// this will result in undefined behavior
func (b *Bucket) ForEach(fn func(k, v []byte) error) error {
	if b.tx.db == nil {
		return ErrTxClosed
	}

	c := b.Cursor()
	for k, v := c.Fir
}
