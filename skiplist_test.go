package skiplist

import (
	"testing"
)

func checkSanity(list *SkipList, t *testing.T) {
	// each level must be correctly ordered
	for k, v := range list.next {
		//t.Log("Level", k)

		if v == nil {
			continue
		}

		if k > len(v.next) {
			t.Fatal("first node's level must be no less than current level")
		}

		next := v
		cnt := 1

		for next.next[k] != nil {
			if !(next.next[k].key >= next.key) {
				t.Fatalf("next key value must be greater than prev key value. [next:%v] [prev:%v]", next.next[k].key, next.key)
			}

			if k > len(next.next) {
				t.Fatalf("node's level must be no less than current level. [cur:%v] [node:%v]", k, next.next)
			}

			next = next.next[k]
			cnt++
		}

		if k == 0 {
			if cnt != list.length {
				t.Fatalf("list len must match the level 0 nodes count. [cur:%v] [level0:%v]", cnt, list.length)
			}
		}
	}
}

func TestBasicIntCRUD(t *testing.T) {
	var list *SkipList

	list = New()

	list.Set(10, 1)
	list.Set(60, 2)
	list.Set(30, 3)
	list.Set(20, 4)
	list.Set(90, 5)
	checkSanity(list, t)

	list.Set(30, 9)
	checkSanity(list, t)

	list.Remove(0)
	list.Remove(20)
	checkSanity(list, t)

	v1 := list.Get(10)
	v2 := list.Get(60)
	v3 := list.Get(30)
	v4 := list.Get(20)
	v5 := list.Get(90)
	v6 := list.Get(0)

	if v1 == nil || v1.value.(int) != 1 || v1.key != 10 {
		t.Fatal(`wrong "10" value (expected "1")`, v1)
	}

	if v2 == nil || v2.value.(int) != 2 {
		t.Fatal(`wrong "60" value (expected "2")`)
	}

	if v3 == nil || v3.value.(int) != 9 {
		t.Fatal(`wrong "30" value (expected "9")`)
	}

	if v4 != nil {
		t.Fatal(`found value for key "20", which should have been deleted`)
	}

	if v5 == nil || v5.value.(int) != 5 {
		t.Fatal(`wrong "90" value`)
	}

	if v6 != nil {
		t.Fatal(`found value for key "0", which should have been deleted`)
	}
}

func TestChangeLevel(t *testing.T) {
	var i uint64

	// Override global default for this test, save old value to restore afterward
	oldMaxLevel := DefaultMaxLevel
	DefaultMaxLevel = 10
	list := New()
	DefaultMaxLevel = oldMaxLevel

	if list.maxLevel != 10 {
		t.Fatal("max level must equal default max value")
	}

	for i = 0; i <= 200; i += 4 {
		list.Set(i, i*10)
	}

	checkSanity(list, t)

	// Test setting the max level just for this list, not the global default
	list.SetMaxLevel(20)
	checkSanity(list, t)

	for i = 1; i <= 201; i += 4 {
		list.Set(i, i*10)
	}

	list.SetMaxLevel(4)
	checkSanity(list, t)

	if list.length != 102 {
		t.Fatal("wrong list length", list.length)
	}

	for c := list.Front(); c != nil; c = c.Next() {
		if c.key*10 != c.value.(uint64) {
			t.Fatal("wrong list element value")
		}
	}
}
