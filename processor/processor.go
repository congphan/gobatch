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
	ErrCanceled         = fmt.Errorf("canceled")
)

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

// Execute process objects by batch by calling funcProcess: batch is a batch of data, batchIndex is index of batch.
// batchIndex will help full whenn you need to convert index of item from current batch to original index of source object.
// error: 1. if objects is not sliceable will return error
// error: 2. error cancel return when it receive cancel singal from context and it will stop for next batch processing
func (p *Processor) Execute(ctx context.Context, objects interface{}, funcProcess func(batch interface{}, batchIndex int)) error {
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
			return ErrCanceled
		default:
			i, j := getIndicesOfNextBatch()
			if i >= length {
				// finish
				return nil
			}

			batch := concreteSliceValue.Slice(i, j)
			if isPointer {
				out := reflect.New(reflect.TypeOf(concreteSliceValue.Interface()))
				out.Elem().Set(batch)
				funcProcess(out.Interface(), nextBatchIdx-1)
			} else {
				funcProcess(batch.Interface(), nextBatchIdx-1)
			}
		}
	}
}

// OriginalIndex help to convert index of item in that batch to original source
func (p *Processor) OriginalIndex(batchNum, itemIdx int) int {
	return (batchNum * p.batchSize) + itemIdx
}
