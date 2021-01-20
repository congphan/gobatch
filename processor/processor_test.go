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
				p, _ := New(2)
				nums := []int{5, 6, 7, 8, 9}
				batchResults := []Batch{}

				ctx, cancel := context.WithTimeout(context.Background(), timeout)
				defer cancel()
				err := p.Execute(ctx, nums, func(batch Batch) {
					batchResults = append(batchResults, batch)
					time.Sleep(timeout * 2)
				})
				assert.EqualError(t, err, context.DeadlineExceeded.Error())
				assert.EqualValues(t, []Batch{{data: []int{5, 6}, index: 0}}, batchResults)
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

	t.Run("OriginalIndex", func(t *testing.T) {
		t.Run("even batch size", func(t *testing.T) {
			p, _ := New(2)
			t.Run("batch 0, index 0", func(t *testing.T) {
				assert.Equal(t, 0, p.OriginalIndex(0, 0))
			})
			t.Run("batch 0, index 1", func(t *testing.T) {
				assert.Equal(t, 1, p.OriginalIndex(0, 1))
			})

			t.Run("batch 1, index 0", func(t *testing.T) {
				assert.Equal(t, 2, p.OriginalIndex(1, 0))
			})
			t.Run("batch 1, index 1", func(t *testing.T) {
				assert.Equal(t, 3, p.OriginalIndex(1, 1))
			})
		})

		t.Run("odd batch size", func(t *testing.T) {
			p, _ := New(3)
			t.Run("batch 0, index 0", func(t *testing.T) {
				assert.Equal(t, 0, p.OriginalIndex(0, 0))
			})
			t.Run("batch 0, index 1", func(t *testing.T) {
				assert.Equal(t, 1, p.OriginalIndex(0, 1))
			})

			t.Run("batch 1, index 0", func(t *testing.T) {
				assert.Equal(t, 3, p.OriginalIndex(1, 0))
			})
			t.Run("batch 1, index 1", func(t *testing.T) {
				assert.Equal(t, 4, p.OriginalIndex(1, 1))
			})
		})
	})
}
