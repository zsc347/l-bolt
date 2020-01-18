package bolt

import "errors"

// These errors can be returned when opening or calling methods on a DB.
var (
	// ErrInvalid is returned when both meta pages on a database are invalid.
	// This typically occurs when a file is not a bolt database.
	ErrInvalid = errors.New("invalid database")

	// ErrVersionMismatch is returned when the data file was created with
	// a different version of Bolt.
	ErrorVersionMismatch = errors.New("version mismatch")

	// ErrChecksum is returned when either meta page checksum does not match
	// or when checksum is 0
	ErrChecksum = errors.New("checksum error")

	// ErrTimeout is returned when a database cannot obtain an exclusive lock
	// on the data file after the timeout passed to Open().
	ErrTimeout = errors.New("timeout")
)

// These errors can occur when beginning or committing a Tx
var (
	// ErrTxNotWritable is returned when performing a write
	// operation on a read-only transaction.
	ErrTxNotWritable = errors.New("tx not writable")

	// ErrTxClosed is returned when committing or rolling back
	// a transaction that has already been committed or rolled back.
	ErrTxClosed = errors.New("tx closed")

	// ErrDatabaseReadonly is returned when a mutating transaction is
	// stared on a read-only database.
	ErrDatabaseReadOnly = errors.New("database is in read-only mode")
)

// These errors can occur when putting or deleting a value or a bucket
var (
	// ErrBucketNotFound is returned when trying to access a bucket that
	// has not been created yet.
	ErrBucketNotFound = errors.New("bucket not found")

	// ErrBucketExists is returned when creating a bucket that already exists.
	ErrBucketExists = errors.New("bucket already exists")

	// ErrBucketNameRequired  is returned when creating a bucket with a blank name
	ErrBucketNameRequied = errors.New("bucket name required")

	// ErrKeyRequired is returned when inserting a zero-length key.
	ErrKeyRequired = errors.New("key required")

	// ErrKeyTooLarge is returned when inserting a key that is larger than MaxKeySize.
	ErrKeyTooLarge = errors.New("key too large")

	// ErrValueTooLarge is returned when inserting a value that is larger than MaxValueSize.
	ErrValueTooLarge = errors.New("value too large")

	// ErrIncompatibleValue is returned when trying create or delete
	// a bucket on an existing non-bucket or when trying to create a
	// delete a non-bucket key on an existing bucket key.
	ErrIncompatibleValue = errors.New("imcompatible value")
)
