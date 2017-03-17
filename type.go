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
	key   float64
	value interface{}
}

type SkipList struct {
	elementNode
	maxLevel       int
	length         int
	randSource     rand.Source
	probability    float64
	probTable      []float64
	mutex          sync.RWMutex
	prevNodesCache []*elementNode
}
