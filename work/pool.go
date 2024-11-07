package work

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Executor interface {
	Execute() error
	OnError(error)
}

type Pool struct {
	numberOfWorkers int
	tasks           chan Executor
	start           sync.Once
	stop            sync.Once
	taskCompleted   chan bool
	quit            chan struct{}
}

func NewPool(noWorkers int, taskChannelSize int) (*Pool, error) {
	if noWorkers <= 0 {
		return nil, errors.New("number of workers cannot be less, or equal to zero")
	}

	if taskChannelSize < 0 {
		return nil, errors.New("channel size cannot be a negative value")
	}

	return &Pool{
		numberOfWorkers: noWorkers,
		tasks:           make(chan Executor, taskChannelSize),
		start:           sync.Once{},
		stop:            sync.Once{},
		taskCompleted:   make(chan bool),
		quit:            make(chan struct{}),
	}, nil
}

func (p *Pool) Start(ctx context.Context) {
	p.start.Do(func() {
		p.startWorker(ctx)
	})
}

func (p *Pool) Stop() {
	p.stop.Do(func() {
		close(p.quit)
	})
}

func (p *Pool) AddTask(t Executor) {
	select {
	case p.tasks <- t:
	case <-p.quit:
	}
}

func (p *Pool) AddNonBlockingTask(t Executor) {
	go func() {
		p.tasks <- t
	}()
}

func (p *Pool) startWorker(ctx context.Context) {
	for i := 0; i < p.numberOfWorkers; i++ {
		go func(workerNum int) {
			fmt.Printf("worker number %d STARTED \n", workerNum)
			for {
				select {
				case <-ctx.Done():
					return
				case <-p.quit:
					return
				case task, ok := <-p.tasks:
					if !ok {
						return
					}

					if err := task.Execute(); err != nil {
						task.OnError(err)
					}

					// make this asynchronously
					// go func() {
					p.taskCompleted <- true
					// }()

					fmt.Printf("worker number %d task FINISHED \n", workerNum)
				}
			}
		}(i)
	}
}

func (p *Pool) TasksCompleted() <-chan bool {
	return p.taskCompleted
}
