package bolt

import "time"

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
	db       *DB
	pages    map[pgid]*page
	stats    TxStats
}

// page returns a reference to the page with a given id.
// If page has been written to then a temporary buffered page is returned
func (tx *Tx) page(id pgid) *page {
	// Check the dirty pages first
	if tx.pages != nil {
		if p, ok := tx.pages[id]; ok {
			return p
		}
	}

	// Otherwise return directly from the mmap
	return tx.db.page(id)
}

// TxStats represents statistics about the actions performed by the transaction.
type TxStats struct {
	// Page statistics
	PageCount int // number of page allocations
	PageAlloc int // total bytes allocated

	// Cursor statistics
	CursorCount int // number of cursors created

	// Node statistics
	NodeCount int // number of node allocations
	NodeDeref int // number of node dereferences

	// Rebalance statictics
	Rebalance     int           // number of node rebalances
	RebalanceTime time.Duration // total time spent rebalancing

	// Split/Spill statistics
	Split     int           // number of nodes split
	Spill     int           // number of nodes spilled
	SpillTime time.Duration // total time spent spilling

	// Write statistics.
	Write     int           // number of writes performed
	WriteTime time.Duration // total time spent writing to disk
}
