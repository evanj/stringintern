package stringintern

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
)

type mapSet struct {
	m      map[string]int32
	values []string
}

type stringSet interface {
	Intern(s string) int
	Index(s string) (int, bool)
	StrValue(i int) (string, bool)
}

func newMapSet() *mapSet {
	return &mapSet{make(map[string]int32), nil}
}

func (s *mapSet) Intern(v string) int {
	index, ok := s.m[v]
	if ok {
		return int(index)
	}

	// does not exist: add it
	index = int32(len(s.values))
	s.values = append(s.values, v)
	s.m[v] = index
	return int(index)
}

func (s *mapSet) Index(v string) (int, bool) {
	index, ok := s.m[v]
	return int(index), ok
}

func (s *mapSet) StrValue(i int) (string, bool) {
	if i < 0 || i >= len(s.values) {
		return "", false
	}
	return s.values[i], true
}

func TestTrivial(t *testing.T) {
	testStrings := []string{"foo", "bar", ""}

	s := New()
	for i, str := range testStrings {
		index, ok := s.Index(str)
		if !(index == 0 && !ok) {
			t.Error(i, str, index, ok)
		}
		v, ok := s.StrValue(i)
		if !(v == "" && !ok) {
			t.Error(i, str, v, ok)
		}

		index = s.Intern(str)
		if index != i {
			t.Errorf("%d: Intern(%#v)=%d; expected %d", i, str, index, i)
		}

		index, ok = s.Index(str)
		if !(index == i && ok) {
			t.Errorf("%d: Index(%#v)=%d, %t; expected %d, %t", i, str, index, ok, i, true)
		}
		v, ok = s.StrValue(i)
		if !(v == str && ok) {
			t.Errorf("%d: StrValue(%d)=%#v, %t; expected %#v, %t", i, i, v, ok, str, true)
		}
	}
}

func strForInt(i int) string {
	return fmt.Sprintf("string%08d", i)
}

func TestImplementations(t *testing.T) {
	const seeds = 10
	const operationsPerSeed = 100000

	for seed := 0; seed < seeds; seed++ {
		implementations := []stringSet{newMapSet(), New()}
		numExisting := 0

		rng := rand.New(rand.NewSource(int64(seed)))
		for op := 0; op < operationsPerSeed; op++ {
			switch rng.Intn(2) {
			case 0:
				// insert new
				newString := strForInt(numExisting)
				expectedIndex := numExisting
				numExisting++

				for _, impl := range implementations {
					v, ok := impl.StrValue(expectedIndex)
					if !(v == "" && !ok) {
						t.Fatal("StrValue of string that should not exist", impl, v, ok, expectedIndex, newString)
					}
					index, ok := impl.Index(newString)
					if !(index == 0 && !ok) {
						t.Fatal("Index of string that should not exist", impl, v, ok)
					}

					index = impl.Intern(newString)
					if index != expectedIndex {
						t.Fatal("Intern of string that should not exist", impl, index, expectedIndex)
					}
					v, ok = impl.StrValue(expectedIndex)
					if !(v == newString && ok) {
						t.Fatal("StrValue of newly added string", impl, v, ok, newString)
					}
					index, ok = impl.Index(newString)
					if !(index == expectedIndex && ok) {
						t.Fatal("Index of newly added string", impl, index, ok, expectedIndex)
					}
					index = impl.Intern(newString)
					if index != expectedIndex {
						t.Fatal("Second intern of string that should not exist", impl, index, expectedIndex)
					}
				}

			case 1:
				if numExisting == 0 {
					break
				}

				// lookup existing string
				expectedIndex := rng.Intn(numExisting)
				existingString := strForInt(expectedIndex)
				for _, impl := range implementations {
					v, ok := impl.StrValue(expectedIndex)
					if !(v == existingString && ok) {
						t.Fatal("Get of existing string", impl, v, ok, expectedIndex, existingString)
					}
					index, ok := impl.Index(existingString)
					if !(index == expectedIndex && ok) {
						t.Fatal("Index of existing string", impl, index, ok, expectedIndex, existingString)
					}
					index = impl.Intern(existingString)
					if !(index == expectedIndex) {
						t.Fatal("Intern of existing string", impl, index, expectedIndex, existingString)
					}
				}

			default:
				panic("BUG: unexpected case")
			}
		}
	}
}

func fill(s stringSet, n int) {
	for i := 0; i < n; i++ {
		s.Intern(strForInt(i))
	}
}

type memMeasure struct {
	before runtime.MemStats
	after  runtime.MemStats
}

func (m *memMeasure) Start() {
	runtime.GC()
	runtime.ReadMemStats(&m.before)
}

func (m *memMeasure) Stop() {
	runtime.GC()
	runtime.ReadMemStats(&m.after)
}

func (m *memMeasure) inUse() int64 {
	return int64(m.after.HeapInuse) - int64(m.before.HeapInuse)
}

func BenchmarkFill(b *testing.B) {
	mem := memMeasure{}

	// pick some "worst case" and "best case" sizes: immediately before/after table resize
	const fullTableSize = (1 << 20) * loadNumerator / loadDenominator
	const emptyTableSize = fullTableSize + 1

	numItems := []int{100, 10000, fullTableSize, emptyTableSize, 10000000}
	for _, items := range numItems {
		b.Run(fmt.Sprintf("map-%d-items", items), func(b *testing.B) {
			mem.Start()
			var m *mapSet

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m = newMapSet()
				fill(m, items)
			}
			b.StopTimer()

			mem.Stop()

			b.ReportMetric(float64(mem.inUse())/float64(items), "B/item")
			if m.Intern("QQQQQ") < 0 {
				panic("ensure map is not GCed")
			}
		})
		b.Run(fmt.Sprintf("intern-%d-items", items), func(b *testing.B) {
			mem.Start()
			var s *Set

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s = New()
				fill(s, items)
			}
			b.StopTimer()

			mem.Stop()
			b.ReportMetric(float64(mem.inUse())/float64(items), "B/item")
			if s.Intern("QQQQQ") < 0 {
				panic("ensure map is not GCed")
			}
		})
	}
}

func BenchmarkIntern(b *testing.B) {
	escape := 0

	// pick some "worst case" and "best case" sizes: immediately before/after table resize
	const fullTableSize = (1 << 20) * loadNumerator / loadDenominator
	const emptyTableSize = fullTableSize + 1

	numItems := []int{100, 10000, fullTableSize, emptyTableSize, 10000000}
	for _, items := range numItems {
		b.Run(fmt.Sprintf("map-%d-items", items), func(b *testing.B) {
			m := newMapSet()
			fill(m, items)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				escape += m.Intern(strForInt(i % items))
			}
		})
		b.Run(fmt.Sprintf("intern-%d-items", items), func(b *testing.B) {
			s := New()
			fill(s, items)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				escape += s.Intern(strForInt(i % items))
			}
		})
	}

	if escape == 0 {
		panic("QQQ escape")
	}
}

// This fake application maps a unique string to 1/10th the number of other strings.
func BenchmarkMapStringToString(b *testing.B) {
	escape := 0
	mem := memMeasure{}

	// there are this many fewer distinct values than keys
	const valueReductionFactor = 10

	// pick some "worst case" and "best case" sizes: immediately before/after table resize
	const fullTableSize = (1 << 20) * loadNumerator / loadDenominator
	const emptyTableSize = fullTableSize + 1

	numItems := []int{100, 10000, fullTableSize, emptyTableSize, 10000000}
	for _, items := range numItems {
		b.Run(fmt.Sprintf("simple-map-%d-items", items), func(b *testing.B) {
			var m map[string]string
			b.Run("fill", func(b *testing.B) {
				mem.Start()
				b.ResetTimer()
				for iteration := 0; iteration < b.N; iteration++ {
					m = make(map[string]string)
					for i := 0; i < items; i++ {
						k := strForInt(i)
						v := strForInt(i / valueReductionFactor)
						m[k] = v
					}
				}
				b.StopTimer()
				mem.Stop()
				b.ReportMetric(float64(mem.inUse())/float64(items), "B/item")
			})

			b.Run("lookup", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					k := strForInt(i)
					// lookup up the value
					v := m[k]
					escape += len(v)
				}
			})
		})

		b.Run(fmt.Sprintf("id-map-%d-items", items), func(b *testing.B) {
			var m map[string]int32
			var idMap *mapSet
			b.Run("fill", func(b *testing.B) {
				mem.Start()
				b.ResetTimer()
				for iteration := 0; iteration < b.N; iteration++ {
					m = make(map[string]int32)
					idMap = newMapSet()
					for i := 0; i < items; i++ {
						k := strForInt(i)
						v := strForInt(i / valueReductionFactor)

						vID := idMap.Intern(v)
						m[k] = int32(vID)
					}
				}
				b.StopTimer()
				mem.Stop()
				b.ReportMetric(float64(mem.inUse())/float64(items), "B/item")
			})

			b.Run("lookup", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					k := strForInt(i)
					// lookup up the value
					vID := m[k]
					v, ok := idMap.StrValue(int(vID))
					if !ok {
						panic("BUG")
					}
					escape += len(v)
				}
			})
		})

		b.Run(fmt.Sprintf("intern-%d-items", items), func(b *testing.B) {
			var m map[string]int32
			var idMap *Set
			b.Run("fill", func(b *testing.B) {
				mem.Start()
				b.ResetTimer()
				for iteration := 0; iteration < b.N; iteration++ {
					m = make(map[string]int32)
					idMap = New()
					for i := 0; i < items; i++ {
						k := strForInt(i)
						v := strForInt(i / valueReductionFactor)

						vID := idMap.Intern(v)
						m[k] = int32(vID)
					}
				}
				b.StopTimer()
				mem.Stop()
				b.ReportMetric(float64(mem.inUse())/float64(items), "B/item")
			})

			b.Run("lookup", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					k := strForInt(i)
					// lookup up the value
					vID := m[k]
					v, ok := idMap.StrValue(int(vID))
					if !ok {
						panic("BUG")
					}
					escape += len(v)
				}
			})
		})
	}

	if escape == 0 {
		panic("QQQ escape")
	}
}
