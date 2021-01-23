package processor_test

import (
	"context"
	"fmt"

	"github.com/congphan/gobatch/processor"
)

func ExampleProcessor_execute() {
	p, _ := processor.New(2)
	p.Execute(context.Background(), []int{5, 6, 7, 8, 9}, func(batch processor.Batch) {
		nums := batch.Data().([]int)
		idx := batch.Index()

		fmt.Println("batch index: ", idx)
		fmt.Println("data: ", nums)
	})
	// Output:
	// batch index:  0
	// data:  [5 6]
	// batch index:  1
	// data:  [7 8]
	// batch index:  2
	// data:  [9]
}

func ExampleProcessor_originalIndex() {
	p, _ := processor.New(2)
	fmt.Println(p.OriginalIndex(1, 1))
	// Output:
	// 3
}
