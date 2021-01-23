package processor_test

import (
	"context"
	"fmt"
	"time"

	"github.com/congphan/gobatch/processor"
)

type employee struct {
	Name string
	Age  int
	Job  string
}

func ExampleProcessor_success() {
	employees := []employee{
		{
			Age:  50,
			Name: "name A1",
			Job:  "JOB 1",
		},
		{
			Age:  60,
			Name: "name A1",
			Job:  "JOB 2",
		},
		{
			Age:  70,
			Name: "ABC XYZ",
			Job:  "JOB 3",
		},
	}

	// simulate function to store list of employee to database
	storeEmployees := func(s []employee) {
		fmt.Println(s)
	}

	p, _ := processor.New(2) // separate employees to batch with maximum 2 employees in a batch if enough
	p.Execute(context.Background(), employees, func(batch processor.Batch) {
		data := batch.Data().([]employee)
		idx := batch.Index()

		fmt.Println("batch index: ", idx)
		storeEmployees(data)
	})
	// Output:
	// batch index:  0
	// [{name A1 50 JOB 1} {name A1 60 JOB 2}]
	// batch index:  1
	// [{ABC XYZ 70 JOB 3}]
}

func ExampleProcessor_cancel() {
	p, _ := processor.New(2)
	nums := []int{5, 6, 7, 8, 9}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // I called cancel() imidiatly to ensure that it always return error cancel
	// In reality cancel() maybe triggerd from another Go routine so if execute finish before trigger than of course Execute() got no error
	err := p.Execute(ctx, nums, func(batch processor.Batch) {
		// your code
	})
	fmt.Println(err) // context.Canceled.
	// OUTPUT:
	// context canceled
}

func ExampleProcessor_timeout() {
	timeout := time.Millisecond * 30
	sleepDuration := timeout * 2
	nums := []int{5, 6, 7, 8, 9}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	p, _ := processor.New(2)
	err := p.Execute(ctx, nums, func(batch processor.Batch) {
		fmt.Println("index: ", batch.Index())
		fmt.Println("data: ", batch.Data())
		time.Sleep(sleepDuration)
	})
	fmt.Println(err) // context.DeadlineExceeded
	// OUTPUT:
	// index:  0
	// data:  [5 6]
	// context deadline exceeded
}

func ExampleProcessor_notSliceable() {
	p, _ := processor.New(1)
	// pass number 1 which is not scliceble
	err := p.Execute(context.Background(), 1, nil)
	fmt.Println(err) // processor.ErrNotSliceable
	// OUTPUT:
	// object not sliceable
}
