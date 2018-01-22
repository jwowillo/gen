package gen_test

import (
	"log"
	"testing"

	"github.com/jwowillo/butler/page"
	"github.com/jwowillo/butler/recipe"
	"github.com/jwowillo/gen"
)

func BenchmarkGen(b *testing.B) {
	rs, err := recipe.List("../butler/book")
	if err != nil {
		log.Fatal(err)
	}
	ps, err := page.List("../butler/web", rs)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		gen.Write("build", gen.AllTransformations, ps)
	}
}