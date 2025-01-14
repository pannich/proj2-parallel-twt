// Package lock provides an implementation of a read-write lock
// that uses condition variables and mutexes.

//An RWMutex (reader/writer mutex) in Go provides a synchronization mechanism
// that allows multiple readers to access a resource concurrently but
// requires exclusive access for writers.
// https://pkg.go.dev/sync#RWMutex

//Basic Components
// Mutex (sync.Mutex): Protects access to the state variables.
// Condition Variable (sync.Cond): Used to manage waiting readers and writers.
// State Variables: Tracks the number of active and waiting readers and writers.

package lock

import (
	"sync"
)

type CustomRWMutex struct {
	mu            *sync.Mutex
	cond          *sync.Cond
	readers       int
	writers       int			// either '1' has active writer or '0' no active writer
	waitingWriters int
	waitingReaders int
}

// return an instance of customRWMutex
func NewCustomRWMutex() *CustomRWMutex {
	var mutex sync.Mutex
	condVar := sync.NewCond(&mutex)		// put the mutex inside the rw instance condition
	return &CustomRWMutex{mu: &mutex, cond: condVar}
}

//Write Lock
// When a write lock is requested, new readers will wait until the writer has released the lock.
func (rw *CustomRWMutex) Lock() {
	// 	Increment the count of waitingWriters.
	// Wait until there are no active readers or writers.
	// Once available, decrement waitingWriters and set writers to 1.

	rw.mu.Lock()					// ensure counter is ++ correctly
	rw.waitingWriters++		 // I request a lock
	for (rw.readers != 0 || rw.writers != 0) {
		// Wait until both condition false. Exit loop when readers == 0 and wrtiers == 0.
		rw.cond.Wait()			// cond.Wait() release the mutex lock.
												// it resume and obtain the lock, when someone signal/broadcast.
	}
	rw.waitingWriters--			// remove myself
	rw.writers = 1					// I get a lock
	rw.mu.Unlock()
}

func (rw *CustomRWMutex) Unlock() {
	// 	Set writers to 0.
	// Broadcast to all waiting readers and writers (to check their conditions).
	rw.mu.Lock()
	rw.writers = 0
	rw.cond.Broadcast()
	rw.mu.Unlock()
}

//Readlock
func (rw *CustomRWMutex) RLock() {
	//	For new reader requesting lock. Wait until there are no writers or waiting writers (to give priority to waiting writers).
	// Increment the count of readers.
	rw.mu.Lock()

	for rw.writers != 0 || rw.waitingWriters != 0 || rw.readers >= 32 {
		// Wait until both condition false. rw.writers == 0 && rw.waitingWriters == 0 && rw.readers < 32.
		rw.cond.Wait()			// resume when someone signal/broadcast
	}
	rw.readers++		 // reader requests a readlock
	rw.mu.Unlock()
}

func (rw *CustomRWMutex) RUnlock() {
	// 	Decrement the count of readers.
	// If the count reaches zero, signal or broadcast to wake waiting writers.
	rw.mu.Lock()

	rw.readers--
	if rw.readers == 0 {
		rw.cond.Broadcast()
	}
	rw.mu.Unlock()
}
