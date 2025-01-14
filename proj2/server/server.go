package server

import (
	"encoding/json"
	"fmt"
	// "io"
	"proj2/queue"
	"proj2/semaphore"
	"sync"
	"sync/atomic"
	"proj2/feed"
)

type Config struct {
	Encoder *json.Encoder // Represents the buffer to encode Responses
	Decoder *json.Decoder // Represents the buffer to decode Requests
	Mode    string        // Represents whether the server should execute
	// sequentially or in parallel
	// If Mode == "s"  then run the sequential version
	// If Mode == "p"  then run the parallel version
	// These are the only values for Version
	ConsumersCount int // Represents the number of consumers to spawn
}

type Response struct {
	Success bool 	`json:"success"` // Maps JSON key "name" to the Go field "Name"
	Id int64 					`json:"id"`
}

type FeedResponse struct {
	Id   int64       `json:"id"`
	Feed []feed.Post `json:"feed"`
}


type SharedContex struct {
	// mu *sync.Mutex
	// cond *sync.Cond
	wgContext *sync.WaitGroup
	sema_con *semaphore.Semaphore	// already internal pointer
	threadCount int
	shutdown int32
}

//Run starts up the twitter server based on the configuration
//information provided and only returns when the server is fully
// shutdown.
func Run(config Config) {
	if config.Mode == "s" {
		RunSequential(config)
	} else if config.Mode == "p" {
		RunParallel(config)
	}
}

func RunSequential(config Config) {
	feed := feed.NewFeed()

	// enqueue
	for {
		var request queue.Request
		err := config.Decoder.Decode(&request)
		if err != nil {
				break 	// EOF
			}
		if request.Command == "DONE" {
			break
		}
		ProcessRequest(&request, feed, config.Encoder)
	}
}

func ProcessRequest(request *queue.Request, feed feed.Feed, encoder *json.Encoder) interface{} {
	// Process the request
	var response interface{}
	switch request.Command {
	case "ADD":
		feed.Add(request.Body, request.Timestamp)
		response = Response{Id: int64(request.ID), Success : true}
	case "REMOVE":
		success := feed.Remove(request.Timestamp)
		response = Response{Id: int64(request.ID), Success: success}
	case "CONTAINS":
		success := feed.Contains(request.Timestamp)
		response = Response{Id: int64(request.ID), Success: success}
	case "FEED":
		posts := feed.Lists()
		response = FeedResponse{Id: int64(request.ID), Feed: posts}
	default:
		response = Response{Id: int64(request.ID), Success : false}
	}

	if err := encoder.Encode(&response); err != nil {
		fmt.Println("Error encoding response:", err)
	}

	return response

}

func RunParallel(config Config) {
	var wg sync.WaitGroup
	var sema_con = semaphore.NewSemaphore(config.ConsumersCount)  // buffer size of 10, starting empty
	// sema_con.Count = 0
	ctx := SharedContex{&wg, sema_con, config.ConsumersCount, 0}

	taskQueue := queue.NewLockFreeQueue()
	feed := feed.NewFeed()

	if config.Mode == "s" {
		Consumer(0, config, taskQueue, &ctx, feed)
	} else if config.Mode == "p" {
		// Consumer Parallel
		for i := 0; i < config.ConsumersCount; i++ {
			ctx.wgContext.Add(1)
			go Consumer(i, config, taskQueue, &ctx, feed)
		}
	}

	Producer(config, &ctx, taskQueue)

	ctx.wgContext.Wait()
}


func Producer (config Config, ctx *SharedContex, taskQueue *queue.LockFreeQueue) {
	for {
		var request queue.Request
		err := config.Decoder.Decode(&request)
		if err != nil {
				return
			}

		if request.Command == "DONE" {
			atomic.StoreInt32(&ctx.shutdown, 1)
			for i:=0; i<ctx.threadCount; i++ {
				ctx.sema_con.Up()
			}
			return
		}

		taskQueue.Enqueue(&request)

		ctx.sema_con.Up()			// 1st call, wake up 1 consumer. If consumer full, wait.
	}
}

// One goRoutine per consumer
func Consumer(goId int, config Config, taskQueue *queue.LockFreeQueue, ctx *SharedContex, feed feed.Feed) {
	defer ctx.wgContext.Done()
	for {
		// free consumers wait here till there's a task
		// wake up and immediately decrement lock
		ctx.sema_con.Down()

		request := taskQueue.Dequeue()
		// fmt.Printf("Consumer %d woke up %v\n", goId, request)

		if request == nil {
			if atomic.LoadInt32(&ctx.shutdown) == 1 {
				return
			}
			continue
		}

		ProcessRequest(request, feed, config.Encoder)

	}
}
