package stringintern

import (
	"github.com/segmentio/fasthash/fnv1a"
)

// increasing the load factor decreases memory usage but increases lookup time.
// At 7/8, the run time is nearly the same as the map, but is much smaller
// 3/4 10M = 42.0 B/item ; 278 ns/op
// 7/8 10M = 40.6 B/item ; 289 ns/op
const loadNumerator = 7
const loadDenominator = 8
const minSize = 16

type Set struct {
	// table maps hash to index plus one so 0 == not present in table
	table  []int32
	values []string
	mask   uint32
}

func New() *Set {
	return &Set{make([]int32, minSize), nil, minSize - 1}
}

// Index returns the integer index for v in the Set, or (0, false) if it is not present.
func (s *Set) Index(v string) (int, bool) {
	slot, found := s.findSlot(v)
	if !found {
		return 0, false
	}

	indexPlusOne := s.table[slot]
	return int(indexPlusOne - 1), true
}

// findSlot returns table index, found.
func (s *Set) findSlot(v string) (int, bool) {
	slot := fnv1a.HashString32(v) & s.mask
	for {
		indexPlusOne := s.table[slot]
		if indexPlusOne == 0 {
			// unused slot: v belongs here
			return int(slot), false
		}
		vIndex := indexPlusOne - 1
		if s.values[vIndex] == v {
			// found the key at slot
			return int(slot), true
		}

		slot = (slot + 1) & s.mask
	}
}

// Intern returns the index for v, adding it if it does not exist.
func (s *Set) Intern(v string) int {
	slot, found := s.findSlot(v)
	if found {
		indexPlusOne := s.table[slot]
		return int(indexPlusOne - 1)
	}

	// add the new string, resizing if necessary
	index := len(s.values)
	maxSize := len(s.table) * loadNumerator / loadDenominator
	// use >= because next num items = index + 1
	if index >= maxSize {
		// resize then search again for the slot
		// fmt.Printf("resizing %d items; max=%d (%d -> %d)\n", index, maxSize, len(s.table), len(s.table)*2)
		s.resize()
		slot, found = s.findSlot(v)
		if found {
			panic("BUG: must not be found after resize")
		}
	}
	s.values = append(s.values, v)
	s.table[slot] = int32(index + 1)
	return index
}

func (s *Set) resize() {
	// discard the old table since we have a list of the keys
	nextSize := len(s.table) * 2
	s.table = make([]int32, nextSize)
	s.mask = uint32(nextSize - 1)

	for i, v := range s.values {
		slot, found := s.findSlot(v)
		if found {
			panic("BUG: must not be found during resize")
		}
		s.table[slot] = int32(i + 1)
	}
}

// Get returns the string corresponding to index, or "", false if it does not exist.
func (s *Set) StrValue(i int) (string, bool) {
	if i < 0 || i >= len(s.values) {
		return "", false
	}
	return s.values[i], true
}
