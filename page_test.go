package bolt

import (
	"reflect"
	"testing"
)

func TestPgids_merge(t *testing.T) {
	a := pgids{4, 5, 6, 10, 11, 12, 13, 27}
	b := pgids{1, 3, 8, 9, 25, 30}
	c := make(pgids, len(a)+len(b))
	merge(c, a, b)
	if !reflect.DeepEqual(c, pgids{1, 3, 4, 5, 6, 8, 9, 10, 11, 12, 13, 25, 27, 30}) {
		t.Errorf("mismatch: %v", c)
	}

	a = pgids{4, 5, 6, 10, 11, 12, 13, 27, 35, 36}
	b = pgids{8, 9, 25, 30}
	c = a.merge(b)
	if !reflect.DeepEqual(c, pgids{4, 5, 6, 8, 9, 10, 11, 12, 13, 25, 27, 30, 35, 36}) {
		t.Errorf("mismatch: %v", c)
	}
}

func BenchmarkPgids_merge(b *testing.B) {
	p1 := pgids{4, 5, 6, 10, 11, 12, 13, 27}
	p2 := pgids{1, 3, 8, 9, 25, 30}

	for i := 0; i < b.N; i++ {
		merged := make(pgids, len(p1)+len(p2))
		mergepgids(merged, p1, p2)
	}
}

func BenchmarkPgids_mergeraw(b *testing.B) {
	p1 := pgids{4, 5, 6, 10, 11, 12, 13, 27}
	p2 := pgids{1, 3, 8, 9, 25, 30}

	for i := 0; i < b.N; i++ {
		merged := make(pgids, len(p1)+len(p2))
		merge(merged, p1, p2)
	}
}
