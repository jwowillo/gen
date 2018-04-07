// Package pipe allows Stages of functions to easily be assembled into concurent
// pipes.
package pipe

import (
	"sync"

	"github.com/jwowillo/gen/page"
)

// Stage handles an page.Page.
type Stage interface {
	Handle(page.Page) page.Page
}

// StageFunc converts a function to a Stage.
type StageFunc func(page.Page) page.Page

// Handle applies the function to the page.Page.
func (f StageFunc) Handle(x page.Page) page.Page {
	return f(x)
}

// Pipe connects Stages so many page.Pages can be processed by the Stages in order
// concurrently.
type Pipe struct {
	m      sync.Mutex
	count  int
	stages []Stage
	links  []chan page.Page
}

// New Pipe with all the Stages connected in the order given.
func New(ss ...Stage) *Pipe {
	return &Pipe{
		m:      sync.Mutex{},
		stages: append([]Stage{}, ss...),
		links:  make([]chan page.Page, len(ss)+1),
	}
}

// Receive the page.Page into the beginning of the Pipe.
func (p *Pipe) Receive(x page.Page) {
	p.m.Lock()
	defer p.m.Unlock()
	if p.isEmpty() {
		p.start()
	}
	p.count++
	go func() { p.links[0] <- x }()
}

// Deliver the item from the end of the Pipe once it's ready.
func (p *Pipe) Deliver() page.Page {
	p.m.Lock()
	defer p.m.Unlock()
	x := <-p.links[len(p.links)-1]
	p.count--
	if p.isEmpty() {
		p.stop()
	}
	return x
}

func (p *Pipe) isEmpty() bool {
	return p.count == 0
}

func (p *Pipe) start() {
	p.links[0] = make(chan page.Page)
	for i, stage := range p.stages {
		p.links[i+1] = make(chan page.Page)
		go func(receive <-chan page.Page, send chan<- page.Page, stage Stage) {
			for x := range receive {
				go func(x page.Page) {
					send <- stage.Handle(x)
				}(x)
			}
		}(p.links[i], p.links[i+1], stage)
	}
}

func (p *Pipe) stop() {
	for _, link := range p.links {
		close(link)
	}
}

// Process all the page.Pages with the Pipe.
func Process(p *Pipe, xs ...page.Page) []page.Page {
	for _, x := range xs {
		p.Receive(x)
	}
	out := make([]page.Page, len(xs))
	for i := range out {
		out[i] = p.Deliver()
	}
	return out
}
