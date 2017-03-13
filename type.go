package skiplist

import (
	"math/rand"
	"sync"
)

type elementNode struct {
	next []*Element
}

type Element struct {
	elementNode
	Value interface{}
	key   uint64
}

// Gets the ajancent next element.
func (element *Element) Next() *Element {
	return element.next[0]
}

// Gets next element at a specific level.
func (element *Element) NextLevel(level int) *Element {
	if level >= len(element.next) || level < 0 {
		panic("invalid argument to NextLevel")
	}

	return element.next[level]
}

// Gets key.
func (element *Element) Key() interface{} {
	return element.key
}

type SkipList struct {
	elementNode
	maxLevel   int
	length     int
	randSource rand.Source
	probTable  []float64
	mutex      sync.RWMutex

	prevNodesCache []*elementNode // a cache for Set/Remove
}
