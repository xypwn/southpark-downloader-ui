package data

import (
	"sort"
	"testing"
)

func TestListBinding(t *testing.T) {
	list := NewListBinding[int]()

	// Test Append
	list.Append(2)
	list.Append(1)
	list.Append(3)

	if list.Len() != 3 {
		t.Errorf("expected list length to be 3, got %d", list.Len())
	}

	// Test Prepend
	list.Prepend(0)
	if list.Get()[0] != 0 {
		t.Errorf("expected first element to be 0, got %d", list.Get()[0])
	}

	// Test Set
	list.Set([]int{5, 4, 3, 2, 1})

	if list.Len() != 5 {
		t.Errorf("expected list length to be 5, got %d", list.Len())
	}

	// Test Sort
	list.Sort(func(a, b int) bool { return a < b })

	if !sort.SliceIsSorted(list.Get(), func(i, j int) bool { return list.Get()[i] < list.Get()[j] }) {
		t.Errorf("expected list to be sorted in ascending order")
	}

	// Test AddListener
	listenerInvokeCount := 0
	listenerID := list.AddListener(func(data []int) {
		listenerInvokeCount++
	})

	list.Append(6)

	if listenerInvokeCount != 2 {
		t.Errorf("expected listener to be invoked twice, but it was invoked %d times", listenerInvokeCount)
	}

	// Test RemoveListener
	list.RemoveListener(listenerID)

	list.Append(8)
	if listenerInvokeCount != 2 {
		t.Errorf("expected listener to remain invoked twice, but it was invoked %d times", listenerInvokeCount)
	}

	// Test Change
	list.Change(func(data []int) []int {
		return append(data, 7)
	})

	if list.Len() != 8 || list.Get()[7] != 7 {
		t.Errorf("expected last element to be 7 and list length to be 8, got last element %d and list length %d", list.Get()[6], list.Len())
	}
}
