package queue

import (
	"sync/atomic"
	"unsafe"
)

type Request struct {
	Command string 	`json:"command"` // Maps JSON key "name" to the Go field "Name"
	ID int 					`json:"id"`
	Body string			`json:"body"`
	Timestamp float64	`json:"timestamp"`
}

// Node represents a node in the queue.
type Node struct {
	value *Request
	next  unsafe.Pointer // *Node
}

// LockFreeQueue represents the queue.
type LockFreeQueue struct {
	head unsafe.Pointer // *Node
	tail unsafe.Pointer // *Node
}

// NewLockFreeQueue creates a new lock-free queue.
func NewLockFreeQueue() *LockFreeQueue {
	dummy := &Node{}
	return &LockFreeQueue{
			head: unsafe.Pointer(dummy),
			tail: unsafe.Pointer(dummy),
	}
}

// Enqueue adds an item to the queue.
func (q *LockFreeQueue) Enqueue(value *Request) {
	newNode := &Node{value: value}
	for {
			tail := atomic.LoadPointer(&q.tail)
			next := atomic.LoadPointer(&(*Node)(tail).next)
			if tail == atomic.LoadPointer(&q.tail) { // Are tail and next consistent?
					if next == nil {
							if atomic.CompareAndSwapPointer(&(*Node)(tail).next, next, unsafe.Pointer(newNode)) {
									atomic.CompareAndSwapPointer(&q.tail, tail, unsafe.Pointer(newNode))
									return
							}
					} else {
							atomic.CompareAndSwapPointer(&q.tail, tail, next)
					}
			}
	}
}

// Dequeue removes and returns an item from the queue.
func (q *LockFreeQueue) Dequeue() *Request {
	for {
			head := atomic.LoadPointer(&q.head)
			tail := atomic.LoadPointer(&q.tail)
			next := atomic.LoadPointer(&(*Node)(head).next)
			if head == atomic.LoadPointer(&q.head) { // Are head, tail, and next consistent?
					if head == tail {
							if next == nil {
									return nil // Queue is empty
							}
							atomic.CompareAndSwapPointer(&q.tail, tail, next)
					} else {
							val := (*Node)(next).value
							if atomic.CompareAndSwapPointer(&q.head, head, next) {
									return val
							}
					}
			}
	}
}

func (q *LockFreeQueue) IsEmpty() bool {
	head := atomic.LoadPointer(&q.head)
	tail := atomic.LoadPointer(&q.tail)
	next := atomic.LoadPointer(&(*Node)(head).next)
	return head == tail && next == nil
}

func (q *LockFreeQueue) Contains(value *Request) bool {
	head := atomic.LoadPointer(&q.head)
	tail := atomic.LoadPointer(&q.tail)
	for curr := (*Node)(head).next; curr != nil && curr != tail; curr = (*Node)(curr).next {
		if (*Node)(curr).value == value {
			return true
		}
	}
	return false
}
