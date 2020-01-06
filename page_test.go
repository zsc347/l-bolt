package bolt

import (
	"testing"
)

func BenchmarkPgids_merge(b *testing.B) {
	p1 := pgids{4, 5, 6, 10, 11, 12, 13, 27}
	p2 := pgids{1, 3, 8, 9, 25, 30}

	for i := 0; i < b.N; i++ {
		merged := make(pgids, len(p1)+len(p2))
		mergepgids(merged, p1, p2)
	}
}
