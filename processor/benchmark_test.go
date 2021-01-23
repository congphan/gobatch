package processor_test

import (
	"context"
	"testing"

	"github.com/congphan/gobatch/processor"
)

var (
	int1Thousand []int = prepareIntegerArray(1000)

	int10Thousand []int = prepareIntegerArray(10000)

	int100Thousand []int = prepareIntegerArray(100000)

	int1Milion []int = prepareIntegerArray(1000000)

	person1Thousand   []person = preparePersonArray(1000)
	person10Thousand  []person = preparePersonArray(10000)
	person100Thousand []person = preparePersonArray(100000)
	person1Milion     []person = preparePersonArray(1000000)
)

type person struct {
	Age  int
	Name string
	Job  string
}

func prepareIntegerArray(size int) []int {
	out := make([]int, size)
	for i := 0; i < size; i++ {
		out[i] = i
	}
	return out
}

func preparePersonArray(size int) []person {
	out := make([]person, size)
	for i := 0; i < size; i++ {
		p := person{
			Age:  70,
			Name: "ABC XYZ",
			Job:  "JOB ABC",
		}
		out[i] = p
	}
	return out
}

func benchmarkProcessorExecute(sources interface{}, batchSize int, f processor.FuncProcess, b *testing.B) {
	for n := 0; n < b.N; n++ {
		p, _ := processor.New(batchSize)
		p.Execute(context.Background(), sources, f)
	}
}

func BenchmarkProcessorExecuteInteger1ThousandElementsBatch100(b *testing.B) {
	benchmarkProcessorExecute(int1Thousand, 100, func(batch processor.Batch) {

	}, b)
}

func BenchmarkProcessorExecuteInteger10ThousandElementsBatch100(b *testing.B) {
	benchmarkProcessorExecute(int10Thousand, 100, func(batch processor.Batch) {

	}, b)
}

func BenchmarkProcessorExecuteInteger100ThousandElementsBatch100(b *testing.B) {
	benchmarkProcessorExecute(int100Thousand, 100, func(batch processor.Batch) {

	}, b)
}

func BenchmarkProcessorExecuteInteger1MilionElementsBatch100(b *testing.B) {
	benchmarkProcessorExecute(int1Milion, 100, func(batch processor.Batch) {

	}, b)
}

func BenchmarkProcessorExecutePersons1ThousandBatch100(b *testing.B) {
	benchmarkProcessorExecute(person1Thousand, 100, func(batch processor.Batch) {

	}, b)
}

func BenchmarkProcessorExecutePersons10ThousandsBatch100(b *testing.B) {
	benchmarkProcessorExecute(person10Thousand, 100, func(batch processor.Batch) {

	}, b)
}

func BenchmarkProcessorExecutePersons100ThousandsBatch100(b *testing.B) {
	benchmarkProcessorExecute(person100Thousand, 100, func(batch processor.Batch) {

	}, b)
}

func BenchmarkProcessorExecutePersons1MilionBatch100(b *testing.B) {
	benchmarkProcessorExecute(person1Milion, 100, func(batch processor.Batch) {

	}, b)
}
