// Package processor define the Processor type, which allow you to execute your given function by batch
package processor

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

var (
	// ErrInvalidBatchSize is returned by the New method if batch size is not positive integer
	ErrInvalidBatchSize = fmt.Errorf("batch size must be positive integer")

	// ErrNotSliceable is returned by Processor.Execute when it receive source object wich is not sliceable
	ErrNotSliceable = fmt.Errorf("object not sliceable")
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

// Processor provide interface to handle data by batch
type Processor interface {
	// Execute process objects by batch by calling funcProcess: batch is a batch of data, batchIndex is index of batch.
	// batchIndex will help ful when you need to convert index of item from current batch to original index of source object.
	//
	// If Execute is not success then a non-nil error return explain why:
	// ErrNotSliceable if objects is not sliceable
	// context.Canceled if the context was canceled and it will stop for next batch processing
	// context.DeadlineExceeded return when the context's deadline passes (timeout)
	// error return from function executed (funcProcess)
	Execute(ctx context.Context, objects interface{}, funcProcess FuncProcess) error
}

// processor struct to execute data by batch
type processor struct {
	m         *sync.Mutex
	batchSize int
}

// New return a Processor which execute data by batch.
//
// ErrInvalidBatchSize returned if batch size is non-positive
func New(batchSize int) (Processor, error) {
	if batchSize < 1 {
		return nil, ErrInvalidBatchSize
	}
	return &processor{
		m:         &sync.Mutex{},
		batchSize: batchSize,
	}, nil
}

// FuncProcess provide signature for function to prcess batch of data
type FuncProcess func(batch Batch) error

func (p *processor) Execute(ctx context.Context, objects interface{}, funcProcess FuncProcess) error {
	executeFuncProcess := func(batch Batch) chan error {
		done := make(chan error)
		go func() {
			done <- funcProcess(batch)
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
			case err := <-chDone: // done for a batch
				if err != nil {
					return err
				}
			}
		}
	}
}
