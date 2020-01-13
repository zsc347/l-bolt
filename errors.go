package bolt

import "errors"

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

	// ErrIncompatibleValue is returned when trying create or delete
	// a bucket on an existing non-bucket or when trying to create a
	// delete a non-bucket key on an existing bucket key.
	ErrIncompatibleValue = errors.New("imcompatible value")
)
