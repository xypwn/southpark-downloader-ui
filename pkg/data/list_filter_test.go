package data

import (
	"reflect"
	"testing"
)

func TestListFilter(t *testing.T) {
	// Setup
	data := NewListBinding[int]()
	filter := NewListFilter(data, func(value int, pattern int) bool {
		return value%pattern == 0
	})
	dataCl := data.NewClient()
	filteredCl := filter.Filtered().NewClient()
	patternCl := filter.Pattern().NewClient()
	var filteredExpect []int

	// Set initial pattern to check div by 5
	patternCl.Change(func(i int) int {
		return 5
	})

	filteredCl.AddListener(func(a []int) {
		if !reflect.DeepEqual(a, filteredExpect) {
			t.Errorf("Expected %v, but got %v", filteredExpect, a)
		}
	})

	// Try changing unfiltered data
	filteredExpect = []int{5, 420, 8000}
	dataCl.Change(func(a []int) []int {
		return []int{1, 2, 3, 4, 5, 69, 420, 8000}
	})

	// Try changing pattern
	filteredExpect = []int{3, 69, 420}
	patternCl.Change(func(i int) int {
		return 3
	})

	// Attempting to change filtered data should panic
	{
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Changing filtered data didn't panic")
			}
		}()

		filteredCl.Change(func(a []int) []int {
			return []int{}
		})
	}
}
