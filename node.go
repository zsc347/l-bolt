package bolt

// node represents an in-memory, deserialized page.
type node struct {
	bucker *Bucket
	isLeaf bool
	inodes inodes
}

// inode represents an internal node inside of a node.
// It can be used to point to elements in a page or point
// to an element which hasn't been added to a page yet.
type inode struct {
	flags uint32
	pgid  pgid
	key   []byte
	value []byte
}

type inodes []inode
