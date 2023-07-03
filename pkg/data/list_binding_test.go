package data

import (
	"sort"
	"testing"
)

func TestListBinding(t *testing.T) {
	list := NewListBinding[int]()
	client1 := list.NewClient()
	client2 := list.NewClient()

	// Test GetCopy, Set and listener function
	listenerInvokeCount := 0
	client2.AddListener(func([]int) {
		listenerInvokeCount++
	})

	client1.Change(func([]int) []int {
		return []int{5, 4, 3, 2, 1}
	})

	client1.Examine(func(arr []int) {
		if len(arr) != 5 {
			t.Errorf("expected list length to be 5, got %d", len(arr))
		}
	})

	if listenerInvokeCount != 1 {
		t.Errorf("expected listener to be invoked once, but it was invoked %d times", listenerInvokeCount)
	}

	// Test Change by sorting
	client1.Change(func(arr []int) []int {
		sort.Slice(arr, func(i, j int) bool {
			return arr[i] < arr[j]
		})
		return arr
	})

	client1.Examine(func(arr []int) {
		if !sort.SliceIsSorted(arr, func(i, j int) bool { return arr[i] < arr[j] }) {
			t.Errorf("expected list to be sorted in ascending order")
		}
	})

	// Test Change and listener invokation
	listenerInvokeCount = 0
	client1.Change(func(arr []int) []int {
		return append(arr, 7)
	})

	client1.Examine(func(arr []int) {
		if len(arr) != 6 || arr[5] != 7 {
			t.Errorf("expected last element to be 7 and list length to be 6, got last element %d and list length %d", arr[5], len(arr))
		}
	})

	if listenerInvokeCount != 1 {
		t.Errorf("expected listener to be invoked once, but it was invoked %d times", listenerInvokeCount)
	}

	// Test listener not invoked when same client changes data
	listenerInvokeCount = 0
	client2.Change(func(arr []int) []int {
		return append(arr, 8)
	})

	if listenerInvokeCount != 0 {
		t.Errorf("expected listener not to be invoked, but it was invoked %d times", listenerInvokeCount)
	}

	// Test RemoveListener
	expectInvok := true
	clientRm := list.NewClient()
	clientRm.AddListener(func([]int) {
		if !expectInvok {
			t.Errorf("listener was invoked after its client was removed")
		}
	})
	client1.Change(func(arr []int) []int {
		return append(arr, 1)
	})
	list.RemoveClient(clientRm)
	expectInvok = false
	client1.Change(func(arr []int) []int {
		return append(arr, 2)
	})
}
