package feed

import (
	// "fmt"
	"proj2/lock"
)

//Feed represents a user's twitter feed
// You will add to this interface the implementations as you complete them.
type Feed interface {
	Add(body string, timestamp float64)
	Remove(timestamp float64) bool
	Contains(timestamp float64) bool
	Lists() []Post
}

type Post struct {
	Body      string
	Timestamp float64
}

//feed is the internal representation of a user's twitter feed (hidden from outside packages)
// You CAN add to this structure but you cannot remove any of the original fields. You must use
// the original fields in your implementation. You can assume the feed will not have duplicate posts
type feed struct {
	start *post // a pointer to the beginning post
	lock	*lock.CustomRWMutex
}

//post is the internal representation of a post on a user's twitter feed (hidden from outside packages)
// You CAN add to this structure but you cannot remove any of the original fields. You must use
// the original fields in your implementation.
type post struct {
	body      string // the text of the post
	timestamp float64  // Unix timestamp of the post
	next      *post  // the next post in the feed
}

//NewPost creates and returns a new post value given its body and timestamp
func NewPost(body string, timestamp float64, next *post) *post {
	return &post{body, timestamp, next}		// like create a new node
}

//NewFeed creates a empy user feed
func NewFeed() Feed {
	newLock := lock.NewCustomRWMutex()
	return &feed{start: &post{}, lock: newLock}							// like a linkedlist with head point to nil
}

// Add inserts a new post to the feed. The feed is always ordered by the timestamp where
// the most recent timestamp is at the beginning of the feed followed by the second most
// recent timestamp, etc. You may need to insert a new post somewhere in the feed because
// the given timestamp may not be the most recent.
func (f *feed) Add(body string, timestamp float64) {
	newP := NewPost(body, timestamp, nil)
	// Lock the feed
	f.lock.Lock()
	defer f.lock.Unlock()

	// Start from the dummy node
	pred := f.start
	curr := pred.next

	// Find the correct insert location
	for curr != nil && curr.timestamp > timestamp {
		pred = curr
		curr = curr.next
	}

	// Insert new post between pred and curr
	newP.next = curr
	pred.next = newP

}

// Remove deletes the post with the given timestamp. If the timestamp
// is not included in a post of the feed then the feed remains
// unchanged. Return true if the deletion was a success, otherwise return false
func (f *feed) Remove(timestamp float64) bool {
	// Lock the feed
	f.lock.Lock()
	defer f.lock.Unlock()

	// Start from the dummy node
	pred := f.start
	curr := pred.next

	// Find the correct insert location
	for curr != nil && curr.timestamp > timestamp {
		pred = curr
		curr = curr.next
	}

	// Delete post between pred and curr
	if curr == nil || curr.timestamp != timestamp {
		return false
	} else {
		pred.next = curr.next
		return true
	}
}

// Contains determines whether a post with the given timestamp is
// inside a feed. The function returns true if there is a post
// with the timestamp, otherwise, false.
func (f *feed) Contains(timestamp float64) bool {
	// Lock the feed
	f.lock.RLock()
	defer f.lock.RUnlock()

	// more than one node
	curr := f.start.next

	for curr != nil {
		if curr.timestamp == timestamp {
			return true
		}
		curr = curr.next
	}
	return false
}

func (f *feed) Lists() []Post {
	f.lock.RLock()
	defer f.lock.RUnlock()

	var posts []Post
	current := f.start.next
	for current != nil {
		posts = append(posts, Post{
			Body:      current.body,
			Timestamp: current.timestamp,
		})
		// fmt.Printf("Post: %v\n", current.body)
		current = current.next
	}
	return posts
}
