package utils

import (
	log "github.com/sirupsen/logrus"
)

type WorkerPool struct {
	queue   chan *job
	workers []*worker
}

type worker struct {
	pool *WorkerPool
	quit chan bool
	id   int
}

type job struct {
	f   func(arg interface{}) error
	a   interface{}
	err chan error
}

func newWorker(pool *WorkerPool, id int) *worker {
	return &worker{pool: pool, quit: make(chan bool, 1), id: id}
}

func (w *worker) run() {
	for {
		select {
		case <-w.quit:
			return
		case job := <-w.pool.queue:
			log.Debugf("Worker %d: received job", w.id)
			err := job.f(job.a)
			job.err <- err
		}
	}
}

func NewWorkerPool(workers int) *WorkerPool {
	pool := &WorkerPool{
		queue: make(chan *job),
	}

	for i := 0; i < workers; i++ {
		pool.workers = append(pool.workers, newWorker(pool, i))
	}

	return pool
}

func (pool *WorkerPool) Start() *WorkerPool {
	for _, worker := range pool.workers {
		go worker.run()
	}

	return pool
}

func (pool *WorkerPool) Stop() {
	for _, worker := range pool.workers {
		worker.quit <- true
	}
}

func (pool *WorkerPool) Call(f func(arg interface{}) error, a interface{}) chan error {
	err := make(chan error, 1)
	pool.queue <- &job{f: f, a: a, err: err}
	return err
}
