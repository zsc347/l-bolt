package bolt

import (
	"fmt"
	"os"
	"unsafe"
)

const pageHeaderSize = int(unsafe.Offsetof(((*page)(nil)).ptr))

const minKeysPerPage = 2

const branchPageElementSize = int(unsafe.Sizeof(branchPageElement{}))
const leafPageElementSize = int(unsafe.Sizeof(leafPageElement{}))

const (
	branchPageFlag   = 0x01
	leafPageFlag     = 0x02
	metaPageFlag     = 0x04
	freelistPageFlag = 0x10
)

const (
	bucketLeafFlag = 0x01
)

type pgid uint64

type page struct {
	id       pgid
	flags    uint16
	count    uint16
	overflow uint32
	ptr      uintptr
}

func (p *page) typ() string {
	if (p.flags & branchPageFlag) != 0 {
		return "branch"
	} else if (p.flags & leafPageFlag) != 0 {
		return "leaf"
	} else if (p.flags & metaPageFlag) != 0 {
		return "meta"
	} else if (p.flags & freelistPageFlag) != 0 {
		return "freelist"
	}
	return fmt.Sprintf("unknown<%2x>", p.flags)
}

func (p *page) meta() *meta {
	return (*meta)(unsafe.Pointer(&p.ptr))
}

func (p *page) leafPageElement(index uint64) *leafPageElement {
	n := &((*[0x7FFFFFF]leafPageElement)(unsafe.Pointer(&p.ptr)))[index]
	return n
}

func (p *page) leafPageElements() []leafPageElement {
	if p.count == 0 {
		return nil
	}
	return ((*[0x7FFFFFF]leafPageElement)(unsafe.Pointer(&p.ptr)))[:]
}

func (p *page) branchPageElement(index uint16) *branchPageElement {
	n := &((*[0x7FFFFFF]branchPageElement)(unsafe.Pointer(&p.ptr)))[index]
	return n
}

func (p *page) branchPageElements() []branchPageElement {
	if p.count == 0 {
		return nil
	}
	return ((*[0x7FFFFFFF]branchPageElement)(unsafe.Pointer(&p.ptr)))[:]
}

func (p *page) hexdump(n int) {
	buf := (*[maxAllocSize]byte)(unsafe.Pointer(p))[:n]
	fmt.Fprintf(os.Stderr, "%x\n", buf)
}

type pages []*page

func (s pages) Len() int           { return len(s) }
func (s pages) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s pages) Less(i, j int) bool { return s[i].id < s[j].id }

type branchPageElement struct {
	pos   uint32
	ksize uint32
	pgid  pgid
}

func (n *branchPageElement) key() []byte {
	buf := (*[maxAllocSize]byte)(unsafe.Pointer(n))
	return (*[maxAllocSize]byte)(unsafe.Pointer(&buf[n.pos]))[:n.ksize:n.ksize]
}

type leafPageElement struct {
	flags uint32
	pos   uint32
	ksize uint32
	vsize uint32
}

func (n *leafPageElement) key() []byte {
	buf := (*[maxAllocSize]byte)(unsafe.Pointer(n))
	return (*[maxAllocSize]byte)(unsafe.Pointer(&buf[n.pos]))[:n.ksize:n.ksize]
}

func (n *leafPageElement) value() []byte {
	buf := (*[maxAllocSize]byte)(unsafe.Pointer(n))
	return (*[maxAllocSize]byte)(unsafe.Pointer(&buf[n.pos+n.ksize]))[:n.vsize:n.vsize]
}

type pgids []pgid

func (s pgids) Len() int           { return len(s) }
func (s pgids) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s pgids) Less(i, j int) bool { return s[i] < s[j] }

func (s pgids) merge(s2 pgids) pgids {
	if len(s) == 0 {
		return s2
	}
	if len(s2) == 0 {
		return s
	}
	merged := make(pgids, len(s)+len(s2))
	mergepgids(merged, s, s2)
	return merged
}

func mergepgids(dst, a, b pgids) {
	la, lb := len(a), len(b)
	if len(dst) < la+lb {
		panic(fmt.Errorf("mergepgids bad len %d < %d + %d", len(dst), len(a), len(b)))
	}

	if la == 0 {
		copy(dst, b)
		return
	}
	if lb == 0 {
		copy(dst, a)
		return
	}

	i, j, k := 0, 0, 0
	for i < la && j < lb {
		if a[i] < b[j] {
			dst[k] = a[i]
			k++
			i++
		} else {
			dst[k] = b[j]
			j++
			k++
		}
	}
	for i < la {
		dst[k] = a[i]
		k++
		i++
	}
	for j < lb {
		dst[k] = b[j]
		j++
		k++
	}
}
