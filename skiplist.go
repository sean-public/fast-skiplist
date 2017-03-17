package skiplist

import (
	"math"
	"math/rand"
	"time"
)

var (
	DefaultMaxLevel    int     = 21
	DefaultProbability float64 = 1 / math.E
)

// Front returns the head node of the list.
func (list *SkipList) Front() *Element {
	return list.next[0]
}

// Next returns the following Element or nil if we're at the end of the list.
// Only operates on the bottom level of the skip list (a fully linked list).
func (element *Element) Next() *Element {
	return element.next[0]
}

// Set inserts a value in the list with the specified key, ordered by the key.
// If the key exists, it updates the value in the existing node.
// Returns a pointer to the new element.
// Locking is optimistic and happens only after searching.
func (list *SkipList) Set(key float64, value interface{}) *Element {
	var element *Element

	prevs := list.getPrevElementNodes(key)

	list.mutex.Lock()
	defer list.mutex.Unlock()

	if element = prevs[0].next[0]; element != nil && element.key <= key {
		element.value = value
		return element
	}

	element = &Element{
		elementNode: elementNode{
			next: make([]*Element, list.randLevel()),
		},
		key:   key,
		value: value,
	}

	for i := range element.next {
		element.next[i] = prevs[i].next[i]
		prevs[i].next[i] = element
	}

	list.length++
	return element
}

// Get finds an element by key. It returns element pointer if found, nil if not found.
// Locking is optimistic and happens only after searching with a fast check for deletion after locking.
func (list *SkipList) Get(key float64) *Element {
	var prev *elementNode = &list.elementNode
	var next *Element

	for i := list.maxLevel - 1; i >= 0; i-- {
		next = prev.next[i]

		for next != nil && key > next.key {
			prev = &next.elementNode
			next = next.next[i]
		}
	}

	list.mutex.Lock()
	defer list.mutex.Unlock()
	if next != nil && next.key <= key {
		return next
	}

	return nil
}

// Remove deletes an element from the list.
// Returns removed element pointer if found, nil if not found.
// Locking is optimistic and happens only after searching with a fast check on adjacent nodes after locking.
func (list *SkipList) Remove(key float64) *Element {
	prevs := list.getPrevElementNodes(key)

	// found the element, remove it
	list.mutex.Lock()
	defer list.mutex.Unlock()
	if element := prevs[0].next[0]; element != nil && element.key <= key {
		for k, v := range element.next {
			prevs[k].next[k] = v
		}

		list.length--
		return element
	}

	return nil
}

// getPrevElementNodes is the private search mechanism that other functions use.
// Finds the previous nodes on each level relative to the current Element and
// caches them. This approach is similar to a "search finger" as described by Pugh:
// http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.17.524
func (list *SkipList) getPrevElementNodes(key float64) []*elementNode {
	var prev *elementNode = &list.elementNode
	var next *Element

	prevs := list.prevNodesCache

	for i := list.maxLevel - 1; i >= 0; i-- {
		next = prev.next[i]

		for next != nil && key > next.key {
			prev = &next.elementNode
			next = next.next[i]
		}

		prevs[i] = prev
	}

	return prevs
}

// SetMaxLevel changes the maximum level in the data structure.
// It doesn't alter any existing data, only sets a limit on future insert heights.
// newLevel must be between 1 and 64 inclusive.
// Returns true if the level was changed.
func (list *SkipList) SetMaxLevel(newLevel int) (ok bool) {
	if 1 > newLevel || newLevel > 64 || newLevel == list.maxLevel {
		return false
	}

	// Downsizing, just truncate the existing data
	if list.maxLevel > newLevel {
		list.next = list.next[:newLevel]
		list.prevNodesCache = list.prevNodesCache[:newLevel]
		list.probTable = probabilityTable(list.probability, newLevel)
		list.maxLevel = newLevel
		return true
	}

	// Upsizing, need to embiggen arrays
	next := make([]*Element, newLevel)
	copy(next, list.next)
	list.next = next
	list.prevNodesCache = make([]*elementNode, newLevel)
	list.probTable = probabilityTable(list.probability, newLevel)
	list.maxLevel = newLevel

	return true
}

// SetProbability changes the current P value of the list.
// It doesn't alter any existing data, only changes how future insert heights are calculated.
func (list *SkipList) SetProbability(newProbability float64) {
	list.probability = newProbability
	list.probTable = probabilityTable(list.probability, list.maxLevel)
}

func (list *SkipList) randLevel() (level int) {
	// Our random number source only has Int63(), so we have to produce a float64 from it
	// Reference: https://golang.org/src/math/rand/rand.go#L150
	r := float64(list.randSource.Int63()) / (1 << 63)

	level = 1
	for level < list.maxLevel && r < list.probTable[level] {
		level++
	}
	return
}

// probabilityTable calculates in advance the probability of a new node having a given level.
// probability is in [0, 1], maxLevel is (0, 64]
// Returns a table of floating point probabilities that each level should be included during an insert.
func probabilityTable(probability float64, maxLevel int) (table []float64) {
	for i := 1; i <= maxLevel; i++ {
		prob := math.Pow(probability, float64(i-1))
		table = append(table, prob)
	}
	return table
}

// New creates a new skip list with default parameters. Returns a pointer to the new list.
func New() *SkipList {
	return &SkipList{
		elementNode:    elementNode{next: make([]*Element, DefaultMaxLevel)},
		prevNodesCache: make([]*elementNode, DefaultMaxLevel),
		maxLevel:       DefaultMaxLevel,

		// Every new list gets its own PRNG source so they don't block one another
		randSource:  rand.New(rand.NewSource(time.Now().UnixNano())),
		probability: DefaultProbability,
		probTable:   probabilityTable(DefaultProbability, DefaultMaxLevel),
	}
}
