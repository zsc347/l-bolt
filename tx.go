package bolt

type txid uint64

// Tx represents a read-only or read/write transaction on the data base.
// Read-only transactions can be used for retriveing values for
// keys and creating cursors.
// Read/write transcations can create and remove buckets and create and
// remove keys.
//
// IMPORTANT: You must comit or rollback transcations when you are done
// with them. Pages can not be reclaimed by the writer until no more
// transcations are using them. A long running read transaction can cause
// the database to quickly frow
type Tx struct {
	writable bool
}
