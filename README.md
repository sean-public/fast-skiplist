##  fast-skiplist



### Purpose

As the basic building block of an in-memory data structure store, I needed an implementation of skip lists in Go. It needed to be easy to use and thread-safe while maintaining the properties of a classic skip list.

There are several skip list implementations in Go. However, they all are implemented in slightly different ways with sparse optimizations and occasional shortcomings. **Please see the [skiplist-survey](https://github.com/sean-public/skiplist-survey) repo for a comparison of Go skip list implementations (including benchmarks).**

The purpose of this repo is to offer a new, fast implementation with an easy-to-use interface that will suit general data storage purposes.

| Operation          | Time Complexity |
| ------------------ | -------- |
| Insertion          | O(log N) |
| Removal            | O(log N) |
| Check if contains  | O(log N) |
| Enumerate in order | O(N)     |


### Quickstart

To start using the library right away, just do:

```sh
go get github.com/sean-public/fast-skiplist
```

There are no external dependencies, so you can start using it right away:

```go
import github.com/sean-public/fast-skiplist

list := skiplist.New()
list.Set(123, "This string data is stored at key 123!")
fmt.Println(list.Get(123).value)
fmt.Println(list.Length)	// prints 1
list.Remove(123)
fmt.Println(list.Length)	// prints 0
```

Of course there are tests, including benchmarks and race condition detection with concurrency:

```
$ go test -cover
PASS
coverage: 100.0% of statements
ok      github.com/sean-public/fast-skiplist    0.006s

$ go test -race
Structure sizes: SkipList is 136, Element is 48 bytes
PASS
ok  	github.com/sean-public/fast-skiplist	41.530s

$ go test -bench=.
Structure sizes: SkipList is 136, Element is 48 bytes
goos: darwin
goarch: amd64
pkg: github.com/sean-public/fast-skiplist
BenchmarkIncSet-8   5000000    370 ns/op    13484040.32 MB/s    62 B/op    3 allocs/op
BenchmarkIncGet-8   10000000   205 ns/op    48592107.58 MB/s    0 B/op     0 allocs/op
BenchmarkDecSet-8   10000000   281 ns/op    35547886.82 MB/s    62 B/op    3 allocs/op
BenchmarkDecGet-8   10000000   212 ns/op    47124462.78 MB/s    0 B/op     0 allocs/op
PASS
ok  	github.com/sean-public/fast-skiplist	21.709s
```



### About fast-skiplist

> "Perfection is achieved not when there is nothing more to add, but rather when there is nothing more to take away"    *— Antoine de Saint-Exupery*

If fast-skiplist is faster than other packages with the same features, it's because it does *less* wherever possible. It locks less, it blocks less, and it traverses less data. Even with these tricks up its sleeve, it has fewer lines of code than most implementations.

###### Calculating the Height of New Nodes

When inserting, it calculates "height" directly instead of consecutive "coin tosses" to add levels. Additionally, it uses a local PRNG source that isn't blocked globally for improved concurrent insert performance.

The probability of adding new nodes to each level of the structure (it's *height*) is determined by successively "rolling the dice" at each level until it doesn't meet a fixed value *P*. The default *P* values for skip lists in the wild range from 0.25 to 0.5. In this implementation, the default is *1/e*, which is optimal for a general-purpose skip list. To find the derivation of this number, see [Analysis of an optimized search algorithm for skip lists](http://www.sciencedirect.com/science/article/pii/030439759400296U) Kirschenhofer et al (1995).

Almost all other implementations are using common functions in `math/rand`, which will block because querying the PRNG to determine the height of new nodes [waits then acquires a lock via the system-wide random number generator](http://blog.sgmansfield.com/2016/01/the-hidden-dangers-of-default-rand/). We get around this by assigning a new rand source to each skip list instantiated, so each skip list can only ever block itself. This significantly speeds up insert times when you are managing multiple lists with high concurrency.

Additionally, this implementation always requests just one number from the PRNG. A pre-computed probability table is used to look up what the *height* of the new node will be. This is faster and offers a fixed calculation time compared to successive "dice rolls" for each level. The table is computed for each level *L* using the default *P* value of *1/e*: `math.Pow(1.0/math.E, L-1)`. It is consulted during inserts by querying for a random number in range [0.0,1.0) and finding the highest level in the table where the random number is less than or equal to the computed number.

For example, let's say `math.Float64()` returned `r=0.029` and the table was pre-computed to contain (with a maximum height of 6):

| height | probability |
| ------ | ----------- |
| 1      | 1.000000000 |
| 2      | 0.367879441 |
| 3      | 0.135335283 |
| 4      | 0.049787068 |
| 5      | 0.018315639 |
| 6      | 0.006737947 |

So the height for the new node would be 5 because *p5 > r ≥ p6*, or 0.018315639 > 0.029 ≥ 0.006737947.

I believe this fast new node height calculation to be novel and faster than any others with user-defined *P* values. [Ticki, for example, proposes an O(1) calculation](http://ticki.github.io/blog/skip-lists-done-right/) but it is fixed to *P=0.5* and I haven't encountered any other optimizations of this calculation. In local benchmarks, this optimization saves 10-25ns per insert.

###### Better Cooperative Multitasking

Why not a lock-free implementation? The overhead created is more than the time spent in contention of a locking version under normal loads. Most research on lock-free structures assume manual alloc/free as well and have separate compaction processes running that are unnecessary in Go (particularly with improved GC as of 1.8). The same is true for the newest variation, [the rotating skip list](http://poseidon.it.usyd.edu.au/~gramoli/web/doc/pubs/rotating-skiplist-preprint-2016.pdf), which claims to be the fastest to date for C/C++ and Java because the compared implementations have maintenance threads with increased overhead for memory management.


###### Caching and Search Fingers

As Pugh described in [A Skip List Cookbook](http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.17.524), search "fingers" can be retained after each lookup operation. When starting the next operation, the finger will point to where the last one occurred and afford the opportunity to pick up the search there instead of starting at the head of the list. This offers *O(log m)* search times, where *m* is the number of elements between the last lookup and the current one (*m* is always less than *n*).

This implementation of a search finger does not suffer the usual problem of "climbing" up in levels when resuming search because it stores pointers to previous nodes for each level independently.



### Benchmarks

Speed is a feature! Below is a set of results demonstrating the flat performance (time per operation) as the list grows to millions of elements. Please see the [skiplist-survey](https://github.com/sean-public/skiplist-survey) repo for complete benchmark results from this and other Go skip list implementations. 

![benchmark results chart](http://i.imgur.com/VqUbsWr.png)



### Todo

- Build more complex test cases (specifically to prove correctness during high concurrency).
- Benchmark memory usage.
- Add "span" to each element to store the distance to the next node on every level. This gives each node a calculable index (ZRANK and associated commands in Redis).
