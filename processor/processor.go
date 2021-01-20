package processor

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// define error constant
var (
	ErrInvalidBatchSize = fmt.Errorf("batch size must be positive integer")
	ErrNotSliceable     = fmt.Errorf("object not sliceable")
)

// Batch is batch of data
type Batch struct {
	data  interface{} // data of batch
	index int         // batch index
}

// Data return data of batch
func (b *Batch) Data() interface{} {
	return b.data
}

// Index return index of batch
func (b *Batch) Index() int {
	return b.index
}

// Processor struct to execute data by batch
type Processor struct {
	m         *sync.Mutex
	batchSize int
}

// New return batch processor
func New(batchSize int) (*Processor, error) {
	if batchSize < 1 {
		return nil, ErrInvalidBatchSize
	}
	return &Processor{
		m:         &sync.Mutex{},
		batchSize: batchSize,
	}, nil
}

// FuncProcess provide signature for function to prcess batch of data
type FuncProcess func(batch Batch)

// Execute process objects by batch by calling funcProcess: batch is a batch of data, batchIndex is index of batch.
// batchIndex will help ful whenn you need to convert index of item from current batch to original index of source object.
// error: 1. if objects is not sliceable will return ErrNotSliceable
// error: 2. error context.Canceled return when it receive cancel singal from context and it will stop for next batch processing
// error: 3. error context.DeadlineExceeded return when the context's deadline passes (timeout)
func (p *Processor) Execute(ctx context.Context, objects interface{}, funcProcess FuncProcess) error {
	executeFuncProcess := func(batch Batch) chan struct{} {
		done := make(chan struct{})
		go func() {
			funcProcess(batch)
			done <- struct{}{}
		}()
		return done
	}

	p.m.Lock()
	defer p.m.Unlock()

	var isPointer bool

	concreteSliceValue := reflect.ValueOf(objects)
	if concreteSliceValue.Kind() == reflect.Ptr {
		concreteSliceValue = concreteSliceValue.Elem()
		isPointer = true
	}

	kind := concreteSliceValue.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return ErrNotSliceable
	}

	length := concreteSliceValue.Len()
	nextStartIdx := 0
	nextBatchIdx := 0
	getIndicesOfNextBatch := func() (i int, j int) {
		i = nextStartIdx
		j = i + p.batchSize
		if j > length {
			j = length
		}
		nextStartIdx = j
		nextBatchIdx++
		return
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			i, j := getIndicesOfNextBatch()
			if i >= length {
				// finish
				return nil
			}

			batchData := concreteSliceValue.Slice(i, j)
			var batch Batch
			if isPointer {
				out := reflect.New(reflect.TypeOf(concreteSliceValue.Interface()))
				out.Elem().Set(batchData)
				batch = Batch{data: out.Interface(), index: nextBatchIdx - 1}
			} else {
				batch = Batch{data: batchData.Interface(), index: nextBatchIdx - 1}
			}

			chDone := executeFuncProcess(batch)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-chDone: // done for a batch
			}
		}
	}
}

// OriginalIndex help to convert index of item in that batch to original source
func (p *Processor) OriginalIndex(batchIndex, itemIndex int) int {
	return (batchIndex * p.batchSize) + itemIndex
}
