package worker

import (
	"errors"
	"net/http"
)

var ErrQueueFull = errors.New("queue is full, backpressure applied")

type Job struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Handler        http.HandlerFunc
	Done           chan struct{}
}

// Pool implements a bounded worker pool for request handling
type Pool struct {
	jobs    chan *Job
	workers int
}

// NewPool initializes a worker pool and starts the worker goroutines.
func NewPool(workers int, queueSize int) *Pool {
	p := &Pool{
		jobs:    make(chan *Job, queueSize),
		workers: workers,
	}
	p.start()
	return p
}

func (p *Pool) start() {
	for i := 0; i < p.workers; i++ {
		go func(workerID int) {
			for job := range p.jobs {
				job.Handler(job.ResponseWriter, job.Request)
				close(job.Done)
			}
		}(i)
	}
}

// Submit tries to enqueue a job. If the queue is full, it immediately returns an error.
func (p *Pool) Submit(w http.ResponseWriter, r *http.Request, h http.HandlerFunc) error {
	job := &Job{
		ResponseWriter: w,
		Request:        r,
		Handler:        h,
		Done:           make(chan struct{}),
	}

	select {
	case p.jobs <- job:
		// Block the client's connection until the worker finishes processing
		<-job.Done
		return nil
	default:
		// Backpressure: queue is full
		return ErrQueueFull
	}
}
