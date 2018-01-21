package main

import (
	"sync"
)

// not thread safe
// FIXME: extract to own module

type Stopper struct {
	ShouldStop chan bool
	numToStop  int
	wg         sync.WaitGroup
}

func NewStopper() *Stopper {
	return &Stopper{
		ShouldStop: make(chan bool, 16),
		numToStop:  0,
		wg:         sync.WaitGroup{},
	}
}

func (s *Stopper) Add() *Stopper {
	s.numToStop++
	s.wg.Add(1)

	return s
}

func (s *Stopper) Done() {
	s.wg.Done()
}

func (s *Stopper) StopAll() {
	for i := 0; i < s.numToStop; i++ {
		s.ShouldStop <- true
	}

	s.wg.Wait()
}
