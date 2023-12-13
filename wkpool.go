package main

import (
	"log"
	"sync"
)

type worker struct {
	id       int
	jobsChan chan func()
	stopChan chan bool
}

func (wk *worker) run(readyWg *sync.WaitGroup) {
	readyWg.Done()

	for {
		select {
		case job := <-wk.jobsChan:
			defer func() {
				if err := recover(); err != nil {
					log.Printf("job #%d panic: %s", wk.id, err)
				}
			}()

			job()
		case <-wk.stopChan:
			return
		}
	}
}

type WkPool struct {
	size     int
	workers  map[int]worker
	jobsChan chan func()
}

func New(size int) *WkPool {
	readyWg := &sync.WaitGroup{}

	pool := &WkPool{
		size:     size,
		workers:  make(map[int]worker),
		jobsChan: make(chan func()),
	}

	readyWg.Add(size)

	for i := 0; i < size; i++ {
		i := i

		wk := worker{
			id:       i,
			jobsChan: pool.jobsChan,
			stopChan: make(chan bool),
		}

		pool.workers[i] = wk
		go wk.run(readyWg)
	}

	readyWg.Wait()

	return pool
}

func (wp *WkPool) Submit(job func()) {
	wp.jobsChan <- job
}

func (wp *WkPool) Stop() {
	for _, wk := range wp.workers {
		wk.stopChan <- true
	}

	close(wp.jobsChan)
}
