package main

import (
	"context"
	"sync"
)

type Task func()

type WorkerPool struct {
	wg           sync.WaitGroup
	workersCount int
	tasks        []Task
	launched     bool
	tasksChan    chan Task
	exit         chan struct{}
}

func NewWorker(workersCount int, tasks []Task) *WorkerPool {
	if workersCount <= 0 {
		panic("workers count must be greater than 0")
	}
	return &WorkerPool{workersCount: workersCount, tasks: tasks, launched: false, tasksChan: make(chan Task, workersCount), exit: make(chan struct{})}
}

func (w *WorkerPool) Run() {
	if w.launched {
		panic("worker pool already launched")
	}
	w.launched = true

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for _, task := range w.tasks {
			w.tasksChan <- task
			select {
			case <-w.exit:
				return
			default:
			}
		}
		close(w.tasksChan)
	}()

	for i := 0; i < w.workersCount; i++ {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			for task := range w.tasksChan {
				task()
				select {
				case <-w.exit:
					return
				default:
				}
			}
		}()
	}

	w.wg.Wait()
}

func (w *WorkerPool) Shutdown(ctx context.Context) error {
	close(w.exit)

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
