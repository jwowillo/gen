package pipe

import (
	"sync"

	"github.com/jwowillo/gen/page"
)

// Consumer consumes page.Pages received from Pipes.
type Consumer interface {
	Consume(page.Page)
}

// ConsumerFunc converts a function to a Consumer.
type ConsumerFunc func(page.Page)

// Consume applies the function to the page.Page.
func (f ConsumerFunc) Consume(x page.Page) {
	f(x)
}

// ProcessAndConsume processes all the page.Pages with the Pipe and gives them to the
// Consumer as they are delivered.
func ProcessAndConsume(p *Pipe, c Consumer, xs ...page.Page) {
	for _, x := range xs {
		p.Receive(x)
	}
	var wg sync.WaitGroup
	wg.Add(len(xs))
	for i := 0; i < len(xs); i++ {
		go func() {
			c.Consume(p.Deliver())
			wg.Done()
		}()
	}
	wg.Wait()
}
