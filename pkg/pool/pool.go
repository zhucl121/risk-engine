// Copyright 2026 RiskEngine Contributors
// SPDX-License-Identifier: Apache-2.0

// Package pool provides a fixed-size goroutine pool for work that must be
// isolated from the Go scheduler (primarily CGO calls like ONNX inference).
//
// CGO calls lock the OS thread they run on, preventing the Go scheduler from
// multiplexing other goroutines onto that thread. Running CGO in a bounded pool
// prevents unbounded OS thread creation and the latency spikes it causes.
package pool

import (
	"context"
	"errors"
)

// ErrPoolFull is returned when the pool's task queue is at capacity.
var ErrPoolFull = errors.New("goroutine pool: queue full")

// Task is a function submitted to the pool.
type Task func()

// Pool runs tasks in a fixed number of worker goroutines.
type Pool struct {
	tasks chan Task
	done  chan struct{}
}

// New creates a Pool with workers goroutines and a task queue of depth queueSize.
// Call Close when done to release resources.
func New(workers, queueSize int) *Pool {
	p := &Pool{
		tasks: make(chan Task, queueSize),
		done:  make(chan struct{}),
	}
	for range workers {
		go p.worker()
	}
	return p
}

func (p *Pool) worker() {
	for {
		select {
		case t := <-p.tasks:
			t()
		case <-p.done:
			return
		}
	}
}

// Submit enqueues a task. It returns ErrPoolFull immediately if the queue is full.
func (p *Pool) Submit(t Task) error {
	select {
	case p.tasks <- t:
		return nil
	default:
		return ErrPoolFull
	}
}

// SubmitWait enqueues a task and blocks until it completes or ctx expires.
func (p *Pool) SubmitWait(ctx context.Context, t Task) error {
	done := make(chan struct{})
	wrapped := func() {
		defer close(done)
		t()
	}
	select {
	case p.tasks <- wrapped:
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close signals all workers to stop. Pending tasks in the queue are abandoned.
func (p *Pool) Close() {
	close(p.done)
}
