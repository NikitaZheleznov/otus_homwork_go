package hw05parallelexecution

import (
	"context"
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	if m <= 0 {
		return ErrErrorsLimitExceeded
	}

	if len(tasks) == 0 {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskCh := make(chan Task)
	errCh := make(chan error)

	var workersWg sync.WaitGroup
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		workersWg.Add(1)
		go worker(ctx, &workersWg, taskCh, errCh)
	}

	wg.Add(1)
	go checkErrors(ctx, &wg, cancel, errCh, m)

	wg.Add(1)
	go addTasks(ctx, &wg, tasks, taskCh)

	workersWg.Wait()
	close(errCh)
	wg.Wait()

	if ctx.Err() != nil {
		return ErrErrorsLimitExceeded
	}

	return nil
}

func worker(ctx context.Context, wg *sync.WaitGroup, taskCh <-chan Task, errCh chan<- error) {
	defer wg.Done()
	for {
		select {
		case task, ok := <-taskCh:
			if !ok {
				return
			}
			if err := task(); err != nil {
				select {
				case errCh <- err:
				case <-ctx.Done():
					return
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func checkErrors(ctx context.Context, wg *sync.WaitGroup, cancel context.CancelFunc, errCh <-chan error, m int) {
	defer wg.Done()
	errorsCount := 0
	for {
		select {
		case err, ok := <-errCh:
			if !ok {
				return
			}
			if err != nil {
				errorsCount++
				if errorsCount >= m {
					cancel()
					return
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func addTasks(ctx context.Context, wg *sync.WaitGroup, tasks []Task, taskCh chan<- Task) {
	defer wg.Done()
	defer close(taskCh)
	for _, task := range tasks {
		select {
		case taskCh <- task:
		case <-ctx.Done():
			return
		}
	}
}
