package data

import (
	"testing"
)

func TestClient(t *testing.T) {
	binding := NewBinding[int]()

	client := binding.NewClient()

	// Initially the value is default to the type
	client.Examine(func(i int) {
		if i != 0 {
			t.Errorf("expected initial value to be 0, got %d", i)
		}
	})

	// Test Change (set)
	client.Change(func(i int) int {
		return 10
	})
	client.Examine(func(i int) {
		if i != 10 {
			t.Errorf("expected value to be 10, got %d", i)
		}
	})

	// Test Change (change)
	client.Change(func(val int) int {
		return val + 10
	})

	client.Examine(func(i int) {
		if i != 20 {
			t.Errorf("expected value to be 20 after Change, got %d", i)
		}
	})

	// Test AddListener
	listenerInvokeCount := 0
	client.AddListener(func(val int) {
		listenerInvokeCount++
	})

	client2 := binding.NewClient()
	client2.Change(func(i int) int {
		return 30
	})

	// Listener function should have been invoked once
	if listenerInvokeCount != 1 {
		t.Errorf("expected listener to be invoked once when value is set by another client, but it was invoked %d times", listenerInvokeCount)
	}

	// Changing value with another client should invoke the listener
	client2.Change(func(val int) int {
		return val + 10
	})

	if listenerInvokeCount != 2 {
		t.Errorf("expected listener to be invoked twice after Change by another client, but it was invoked %d times", listenerInvokeCount)
	}

	// Test RemoveListener
	expectInvok := true
	clientRm := binding.NewClient()
	clientRm.AddListener(func(i int) {
		if !expectInvok {
			t.Errorf("listener was invoked after its client was removed")
		}
	})
	client.Change(func(int) int {
		return 69
	})
	binding.RemoveClient(clientRm)
	expectInvok = false
	client.Change(func(int) int {
		return 420
	})
}
