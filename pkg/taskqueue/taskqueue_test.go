package taskqueue

/*import (
	"context"
	"testing"
	"time"
)

func TestSemaphoreBasic(t *testing.T) {
	s := New(2)

	for i := 0; i < 5; i++ {
		go func() {
			err := s.Acquire(context.Background(), 0, func(h Handle) {})
			if err != nil {
				t.Errorf("error acquiring semaphore: %v", err)
			}

			time.Sleep(10 * time.Millisecond)

			s.Release()
		}()
	}

	time.Sleep(100 * time.Millisecond)
}

func TestSemaphorePriority(t *testing.T) {
	s := New(2)

	prios := [6]int{0, 1, 2, -1, 4, -69}
	done := make(chan int, 6)
	for i := 0; i < 6; i++ {
		index := i
		priority := prios[i]
		time.Sleep(5 * time.Millisecond)
		go func() {
			err := s.Acquire(context.Background(), priority, func(h Handle) {})
			if err != nil {
				t.Errorf("error acquiring semaphore: %v", err)
			}

			time.Sleep(50 * time.Millisecond)

			s.Release()

			done <- index
		}()
	}

	time.Sleep(400 * time.Millisecond)

	close(done)

	expectation := [6]int{
		// Explanations:
		0, // First added -> immediately acquires
		1, // Second added -> immediately acquires
		5, // Lowest prio (-69)
		3, // Next lowest prio (-1)
		2, // Next lowest prio (2)
		4, // Next lowest prio (4)
	}
	result := []int{}
	for v := range done {
		result = append(result, v)
	}

	for i := range expectation {
		if result[i] != expectation[i] {
			t.Fatalf("unexpected order: want %v, got %v", expectation, result)
		}
	}
}

func TestSemaphoreDecreaseSize(t *testing.T) {
	s := New(2)

	prios := [6]int{0, 1, 2, -1, 4, -69}
	done := make(chan int, 6)
	for i := 0; i < 6; i++ {
		index := i
		priority := prios[i]
		time.Sleep(5 * time.Millisecond)
		go func() {
			err := s.Acquire(context.Background(), priority, func(h Handle) {})
			if err != nil {
				t.Errorf("error acquiring semaphore: %v", err)
			}

			time.Sleep(50 * time.Millisecond)

			s.Release()

			done <- index
		}()

		if i == 4 {
			time.Sleep(3 * time.Millisecond)
			s.SetSize(1)
		}
	}

	time.Sleep(400 * time.Millisecond)

	close(done)

	// Order should be no different if semaphore is shrunk
	expectation := [6]int{
		// Explanations: See TestSemaphorePriority
		0, 1, 5, 3, 2, 4,
	}
	result := []int{}
	for v := range done {
		result = append(result, v)
	}

	for i := range expectation {
		if result[i] != expectation[i] {
			t.Fatalf("unexpected order: want %v, got %v", expectation, result)
		}
	}
}

func TestSemaphoreIncreaseSize(t *testing.T) {
	s := New(2)

	prios := [6]int{0, 1, 2, -1, 4, -69}
	done := make(chan int, 6)
	for i := 0; i < 6; i++ {
		index := i
		priority := prios[i]
		time.Sleep(5 * time.Millisecond)
		go func() {
			err := s.Acquire(context.Background(), priority, func(h Handle) {})
			if err != nil {
				t.Errorf("error acquiring semaphore: %v", err)
			}

			time.Sleep(50 * time.Millisecond)

			s.Release()

			done <- index
		}()

		if i == 4 {
			time.Sleep(3 * time.Millisecond)
			s.SetSize(3)
		}
	}

	time.Sleep(400 * time.Millisecond)

	close(done)

	expectation := [6]int{
		// Explanations:
		0, // First added -> immediately acquires
		1, // Second added -> immediately acquires
		3, // When resized to three, known prios are 0, 1, 2, -1, of which -1 is lowest
		5, // Lowest prio (-69)
		2, // Next lowest prio (2)
		4, // Next lowest prio (4)
	}
	result := []int{}
	for v := range done {
		result = append(result, v)
	}

	for i := range expectation {
		if result[i] != expectation[i] {
			t.Fatalf("unexpected order: want %v, got %v", expectation, result)
		}
	}
}
*/
