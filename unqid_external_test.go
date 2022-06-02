package unqid_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/shivakar/unqid"
)

func BenchmarkUnqidID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = unqid.Next()
	}
}

func ExampleNext() {
	n := 100
	ids := make([]int64, 0, 100)
	for i := 0; i < n; i++ {
		id := unqid.Next()
		ids = append(ids, id)
		time.Sleep(2 * time.Millisecond)
	}

	for _, id := range ids {
		if id <= 0 {
			fmt.Println(id)
		}
	}
	// Output:
}
