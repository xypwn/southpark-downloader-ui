package data

import (
	"testing"
)

func TestBinding(t *testing.T) {
	binding := NewBinding[int]()

	// Initially the value is default to the type
	if binding.Get() != 0 {
		t.Errorf("expected initial value to be 0, got %d", binding.Get())
	}

	// Test Set
	binding.Set(10)
	if binding.Get() != 10 {
		t.Errorf("expected value to be 10, got %d", binding.Get())
	}

	// Test Change
	binding.Change(func(val int) int {
		return val + 10
	})

	if binding.Get() != 20 {
		t.Errorf("expected value to be 20 after Change, got %d", binding.Get())
	}

	// Test AddListener
	listenerInvokeCount := 0
	listenerID := binding.AddListener(func(val int) {
		listenerInvokeCount++
	})

	if listenerInvokeCount != 1 {
		t.Errorf("expected listener to be invoked once when added, but it was invoked %d times", listenerInvokeCount)
	}

	// Changing value should invoke the listener
	binding.Change(func(val int) int {
		return val + 10
	})

	if listenerInvokeCount != 2 {
		t.Errorf("expected listener to be invoked twice after Change, but it was invoked %d times", listenerInvokeCount)
	}

	// Test RemoveListener
	binding.RemoveListener(listenerID)

	binding.Change(func(val int) int {
		return val + 10
	})

	if listenerInvokeCount != 2 {
		t.Errorf("expected listener to remain invoked twice after RemoveListener, but it was invoked %d times", listenerInvokeCount)
	}
}
