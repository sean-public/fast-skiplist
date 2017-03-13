package skiplist

import (
	"math"
	"math/rand"
	"time"
)

var (
	DefaultMaxLevel int = 21
)

// probabilityTable calculates in advance the probability of a new node having a given level
func probabilityTable(maxLevel int) (table []float64) {
	for i := 1; i <= maxLevel; i++ {
		prob := math.Pow(1.0/math.E, float64(i-1))
		table = append(table, prob)
	}
	return table
}

func (list *SkipList) randLevel() (level int) {
	// Our random number source only has Int63(), so we have to produce a float64 from it
	// Reference: https://golang.org/src/math/rand/rand.go#L150
	r := float64(list.randSource.Int63()) / (1 << 63)

	for level < list.maxLevel && r < list.probTable[level] {
		level++
	}
	return
}

// New creates a new skip list with default parameters. Returns a pointer to the new list.
func New() *SkipList {
	return &SkipList{
		elementNode:    elementNode{next: make([]*Element, DefaultMaxLevel)},
		prevNodesCache: make([]*elementNode, DefaultMaxLevel),
		maxLevel:       DefaultMaxLevel,

		// Every new list gets its own PRNG source so they don't block one another
		randSource: rand.New(rand.NewSource(time.Now().UnixNano())),
		probTable:  probabilityTable(DefaultMaxLevel),
	}
}

// Front returns the head node of the list.
func (list *SkipList) Front() *Element {
	return list.next[0]
}

// Set inserts a value in the list with the specified key, ordered by the key.
// If the key exists, it updates the value in the existing node.
// Returns a pointer to the new element.
// Locking is optimistic and happens only after searching.
func (list *SkipList) Set(key uint64, value interface{}) *Element {
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
func (list *SkipList) Get(key uint64) *Element {
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

// Remove deletes an element form the list.
// Returns removed element pointer if found, nil if not found.
// Locking is optimistic and happens only after searching with a fast check on adjacent nodes after locking.
func (list *SkipList) Remove(key uint64) *Element {
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
func (list *SkipList) getPrevElementNodes(key uint64) []*elementNode {
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

// SetMaxLevel changes skip list max level.
// It doesn't change any nodes already inserted.
func (list *SkipList) SetMaxLevel(level int) (old int) {
	old, list.maxLevel = list.maxLevel, level

	if old == level {
		return
	}

	if old > level {
		list.next = list.next[:level]
		list.prevNodesCache = list.prevNodesCache[:level]
		return
	}

	next := make([]*Element, level)
	copy(next, list.next)
	list.next = next
	list.prevNodesCache = make([]*elementNode, level)
	list.probTable = probabilityTable(level)

	return
}
