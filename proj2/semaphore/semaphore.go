package semaphore

import (
	"sync"
)

type Semaphore struct {
	Count int           // Current value of the semaphore
	mu    sync.Mutex    // Mutex to protect changes to the count
	cond  *sync.Cond    // Condition variable to signal changes in availability
}

// NewSemaphore creates a new semaphore with the given initial and maximum
func NewSemaphore(max int) *Semaphore {
	sem := &Semaphore{
		Count: max,
	}
	sem.cond = sync.NewCond(&sem.mu)
	return sem
}

func (s *Semaphore) Up() {
	s.mu.Lock()
	s.Count++
	s.cond.Signal()
	s.mu.Unlock()
}

func (s *Semaphore) Down() {
	s.mu.Lock()
	for s.Count == 0 {
		s.cond.Wait()
	}
	s.Count--
	s.mu.Unlock()
}
