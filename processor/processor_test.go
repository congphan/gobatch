package processor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcessor_Execute(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		t.Run("slice", func(t *testing.T) {
			p, _ := New(1)
			var nilSlice []int = nil
			err := p.Execute(context.Background(), nilSlice, nil)
			assert.NoError(t, err)
		})

		t.Run("another type", func(t *testing.T) {
			p, _ := New(1)
			var nilInt *int = nil
			err := p.Execute(context.Background(), nilInt, nil)
			assert.EqualError(t, err, ErrNotSliceable.Error())
		})
	})

	t.Run("invalid batch size", func(t *testing.T) {
		_, err := New(0)
		assert.EqualError(t, err, ErrInvalidBatchSize.Error())
	})

	t.Run("not slice", func(t *testing.T) {
		p, _ := New(1)
		err := p.Execute(context.Background(), 1, nil)
		assert.EqualError(t, err, ErrNotSliceable.Error())
	})

	t.Run("success", func(t *testing.T) {
		t.Run("slice", func(t *testing.T) {
			t.Run("last batch is less then batch size", func(t *testing.T) {
				p, _ := New(2)
				nums := []int{5, 6, 7, 8, 9}
				batchResults := []Batch{}
				err := p.Execute(context.Background(), nums, func(batch Batch) {
					batchResults = append(batchResults, batch)
				})
				assert.NoError(t, err)
				assert.EqualValues(t, []Batch{
					{
						data:  []int{5, 6},
						index: 0,
					},
					{
						data:  []int{7, 8},
						index: 1,
					},
					{
						data:  []int{9},
						index: 2,
					},
				}, batchResults)
			})

			t.Run("last batch is equal batch size", func(t *testing.T) {
				p, _ := New(2)
				nums := []int{5, 6, 7, 8}
				batchResults := []Batch{}
				err := p.Execute(context.Background(), nums, func(batch Batch) {
					batchResults = append(batchResults, batch)
				})
				assert.NoError(t, err)
				assert.EqualValues(t, []Batch{
					{
						data:  []int{5, 6},
						index: 0,
					},
					{
						data:  []int{7, 8},
						index: 1,
					},
				}, batchResults)
			})

			t.Run("empty data", func(t *testing.T) {
				p, _ := New(2)
				nums := []int{}
				batchResults := []Batch{}
				err := p.Execute(context.Background(), nums, func(batch Batch) {
					batchResults = append(batchResults, batch)
				})
				assert.NoError(t, err)
				assert.EqualValues(t, []Batch{}, batchResults)
			})

			t.Run("cancel", func(t *testing.T) {
				p, _ := New(2)
				nums := []int{5, 6, 7, 8, 9}
				batchResults := []Batch{}
				err := p.Execute(func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(), nums, func(batch Batch) {
					batchResults = append(batchResults, batch)
				})
				assert.EqualError(t, err, context.Canceled.Error())
				assert.EqualValues(t, []Batch{}, batchResults)
			})

			t.Run("timeout", func(t *testing.T) {
				timeout := time.Millisecond * 30
				sleepDuration := timeout * 2
				p, _ := New(2)
				nums := []int{5, 6, 7, 8, 9}
				batchResults := []Batch{}

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()

				start := time.Now()
				err := p.Execute(ctx, nums, func(batch Batch) {
					batchResults = append(batchResults, batch)
					time.Sleep(sleepDuration)
				})
				taken := time.Since(start)
				assert.EqualError(t, err, context.DeadlineExceeded.Error())
				assert.EqualValues(t, []Batch{{data: []int{5, 6}, index: 0}}, batchResults)
				assert.True(t, taken.Milliseconds() < sleepDuration.Milliseconds()) // expected time executed must lest then sleep duration
			})
		})

		t.Run("pointer of slice", func(t *testing.T) {
			p, _ := New(2)
			nums := &[]int{5, 6, 7, 8, 9}
			batchResults := []Batch{}
			err := p.Execute(context.Background(), nums, func(batch Batch) {
				batchResults = append(batchResults, batch)
			})
			assert.NoError(t, err)
			assert.EqualValues(t, []Batch{
				{
					data:  &[]int{5, 6},
					index: 0,
				},
				{
					data:  &[]int{7, 8},
					index: 1,
				},
				{
					data:  &[]int{9},
					index: 2,
				},
			}, batchResults)
		})
	})
}
