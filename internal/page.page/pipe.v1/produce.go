package pipe

import (
	"sync"

	"github.com/jwowillo/gen/page"
)

// Producer produces page.Pages to pass to Pipes.
//
// false is returned when the Producer is done producing. The page.Page returned
// with the false value is ignored.
type Producer interface {
	Produce() (page.Page, bool)
}

// ProducerFunc converts a function to a Producer.
type ProducerFunc func() (page.Page, bool)

// Produce applies to the function.
func (f ProducerFunc) Produce() (page.Page, bool) {
	return f()
}

// ProduceAndProcess processes all the page.Pages from the Producer with the Pipe.
func ProduceAndProcess(p *Pipe, prod Producer) []page.Page {
	n := 0
	x, ok := prod.Produce()
	for ok {
		n++
		p.Receive(x)
		x, ok = prod.Produce()
	}
	xs := make([]page.Page, n)
	for i := 0; i < n; i++ {
		xs[i] = p.Deliver()
	}
	return xs
}

// ProduceProcessAndConsume process all the page.Pages from the Producer with the
// Pipe and gives them to the Consumer as they are delivered.
func ProduceProcessAndConsume(p *Pipe, prod Producer, c Consumer) {
	n := 0
	x, ok := prod.Produce()
	for ok {
		n++
		p.Receive(x)
		x, ok = prod.Produce()
	}
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			c.Consume(p.Deliver())
			wg.Done()
		}()
	}
	wg.Wait()
}
