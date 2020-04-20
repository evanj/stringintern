# Saving Space by Mapping Big Objects to Small Integer IDs

This is an experiment in using a custom hash table implementation to reduce memory space. See my blog post for a detailed explanation: https://www.evanjones.ca/saving-memory-dense-integers.html

If you want to use this implementation, you should probably copy it into your application. Due to Go's lack of generics, if you want to use it with a type that is not a string, then you should just search/replace the types.


## Random implementation notes

* The custom hash table uses a load factor of 7/8. This made the custom table lookups approximately the same speed as the standard library hash table, while taking less storage.


## Benchmarks

From `go test -bench=. -benchmem -benchtime=2s .`. This shows that the "custom" implementation is around the same speed to fill and lookup as the implementation that uses the standard library map, but occupies a fair amount of less space. When used to replace a `map[string]string`, it takes longer to fill the map, but is actually faster to look up. This must be due to caching?

* id-map = The Go standard `map` plus an array to map strings to IDs
* intern = custom hash table plus an array to map strings to IDs
* simple-map = A plain `map[string]string`
* BenchmarkFill = Time to fill a string to ID table with X unique strings.
* BenchmarkIntern = Time to lookup a string and get the matching ID.
* BenchmarkStringToString = Time to fill and look up a `map[string]string` with and without an ID mapping.

```
BenchmarkFill/map-100-items-4                83474       28557 ns/op         983 B/item        13062 B/op       219 allocs/op
BenchmarkFill/intern-100-items-4            109723       21762 ns/op         492 B/item         7497 B/op       212 allocs/op
BenchmarkFill/map-10000-items-4                770     3248146 ns/op          87.7 B/item    1887512 B/op     20281 allocs/op
BenchmarkFill/intern-10000-items-4             800     2809862 ns/op          63.9 B/item    1197266 B/op     20032 allocs/op
BenchmarkFill/map-917504-items-4                 5   484156368 ns/op         122 B/item    215313550 B/op   1873316 allocs/op
BenchmarkFill/intern-917504-items-4              5   507906036 ns/op          44.6 B/item  118916851 B/op   1835085 allocs/op
BenchmarkFill/map-917505-items-4                 4   504956605 ns/op         122 B/item    215340850 B/op   1873472 allocs/op
BenchmarkFill/intern-917505-items-4              4   551814650 ns/op          46.3 B/item  127307282 B/op   1835102 allocs/op
BenchmarkFill/map-10000000-items-4               1  7189183496 ns/op          74.3 B/item 1905868608 B/op  20306834 allocs/op
BenchmarkFill/intern-10000000-items-4            1  7304909496 ns/op          40.6 B/item 1201760568 B/op  20000166 allocs/op
BenchmarkIntern/map-100-items-4           12228962         197 ns/op        23 B/op        1 allocs/op
BenchmarkIntern/intern-100-items-4        12440253         203 ns/op        23 B/op        1 allocs/op
BenchmarkIntern/map-10000-items-4         11120804         217 ns/op        24 B/op        1 allocs/op
BenchmarkIntern/intern-10000-items-4      11511997         210 ns/op        24 B/op        1 allocs/op
BenchmarkIntern/map-917504-items-4         6994010         328 ns/op        24 B/op        2 allocs/op
BenchmarkIntern/intern-917504-items-4      6867302         369 ns/op        24 B/op        2 allocs/op
BenchmarkIntern/map-917505-items-4         7868085         319 ns/op        24 B/op        1 allocs/op
BenchmarkIntern/intern-917505-items-4      9460183         259 ns/op        24 B/op        2 allocs/op
BenchmarkIntern/map-10000000-items-4       8141082         319 ns/op        24 B/op        1 allocs/op
BenchmarkIntern/intern-10000000-items-4    8883363         332 ns/op        23 B/op        1 allocs/op
BenchmarkMapStringToString/simple-map-100-items/fill-4         58017       43146 ns/op         328 B/item    15226 B/op      400 allocs/op
BenchmarkMapStringToString/simple-map-100-items/lookup-4    12616658         192 ns/op        24 B/op        2 allocs/op
BenchmarkMapStringToString/id-map-100-items/fill-4             55298       43486 ns/op         410 B/item    12377 B/op      409 allocs/op
BenchmarkMapStringToString/id-map-100-items/lookup-4        12301394         198 ns/op        24 B/op        2 allocs/op
BenchmarkMapStringToString/intern-100-items/fill-4             56949       40999 ns/op         246 B/item    11894 B/op      407 allocs/op
BenchmarkMapStringToString/intern-100-items/lookup-4        12544710         194 ns/op        24 B/op        2 allocs/op
BenchmarkMapStringToString/simple-map-10000-items/fill-4         538     4697703 ns/op           1.64 B/item   1755388 B/op    40205 allocs/op
BenchmarkMapStringToString/simple-map-10000-items/lookup-4          11188522         223 ns/op        24 B/op        2 allocs/op
BenchmarkMapStringToString/id-map-10000-items/fill-4                     524     4585083 ns/op          16.4 B/item  1438014 B/op    40307 allocs/op
BenchmarkMapStringToString/id-map-10000-items/lookup-4              11947944         203 ns/op        24 B/op        2 allocs/op
BenchmarkMapStringToString/intern-10000-items/fill-4                     520     4614346 ns/op           6.55 B/item   1350303 B/op    40269 allocs/op
BenchmarkMapStringToString/intern-10000-items/lookup-4              12020376         201 ns/op        24 B/op        2 allocs/op
BenchmarkMapStringToString/simple-map-917504-items/fill-4                  4   619430539 ns/op           0.571 B/item 206601490 B/op   3708241 allocs/op
BenchmarkMapStringToString/simple-map-917504-items/lookup-4          7707883         324 ns/op        24 B/op        2 allocs/op
BenchmarkMapStringToString/id-map-917504-items/fill-4                      4   738044341 ns/op           0.446 B/item 162891434 B/op   3711570 allocs/op
BenchmarkMapStringToString/id-map-917504-items/lookup-4              7992980         319 ns/op        24 B/op        2 allocs/op
BenchmarkMapStringToString/intern-917504-items/fill-4                      3   695077298 ns/op          13.4 B/item 157250962 B/op   3708467 allocs/op
BenchmarkMapStringToString/intern-917504-items/lookup-4              7163472         303 ns/op        24 B/op        2 allocs/op
BenchmarkMapStringToString/simple-map-917505-items/fill-4                  4   620762208 ns/op          -1.15 B/item  206600938 B/op   3708241 allocs/op
BenchmarkMapStringToString/simple-map-917505-items/lookup-4          7820898         318 ns/op        24 B/op        2 allocs/op
BenchmarkMapStringToString/id-map-917505-items/fill-4                      4   637765959 ns/op           0.116 B/item 162891282 B/op   3711572 allocs/op
BenchmarkMapStringToString/id-map-917505-items/lookup-4              8178648         304 ns/op        24 B/op        2 allocs/op
BenchmarkMapStringToString/intern-917505-items/fill-4                      3   714617491 ns/op          13.5 B/item 157226968 B/op   3708335 allocs/op
BenchmarkMapStringToString/intern-917505-items/lookup-4              8057608         321 ns/op        24 B/op        2 allocs/op
BenchmarkMapStringToString/simple-map-10000000-items/fill-4                1  8156199350 ns/op          97.7 B/item 1780599000 B/op 40307012 allocs/op
BenchmarkMapStringToString/simple-map-10000000-items/lookup-4        7003670         331 ns/op        24 B/op        2 allocs/op
BenchmarkMapStringToString/id-map-10000000-items/fill-4                    1  8100420730 ns/op          73.2 B/item 1511788896 B/op 40345973 allocs/op
BenchmarkMapStringToString/id-map-10000000-items/lookup-4            7610646         312 ns/op        24 B/op        2 allocs/op
BenchmarkMapStringToString/intern-10000000-items/fill-4                    1  8753852649 ns/op          69.1 B/item 1423459560 B/op 40305912 allocs/op
BenchmarkMapStringToString/intern-10000000-items/lookup-4            7750686         331 ns/op        24 B/op        2 allocs/op
```
