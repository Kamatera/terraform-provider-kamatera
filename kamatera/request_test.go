package kamatera

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_changeDisks(t *testing.T) {
	skipWaiting = true
	defer func() {
		skipWaiting = false
	}()

	tests := []struct {
		name     string
		op       diskOperation
		expected []changeDisksPostValues
	}{
		{
			name: "add only",
			op:   diskOperation{add: []float64{10}},
			expected: []changeDisksPostValues{
				{
					ID:  "1",
					Add: "10gb",
				},
			},
		},
		{
			name: "remove only",
			op:   diskOperation{remove: []int{1}},
			expected: []changeDisksPostValues{
				{
					ID:     "1",
					Remove: "1",
				},
			},
		},
		{
			name: "update only",
			op:   diskOperation{update: map[int]float64{1: 10}},
			expected: []changeDisksPostValues{
				{
					ID:     "1",
					Resize: "1",
					Size:   "10gb",
				},
			},
		},
		{
			name: "update and add",
			op: diskOperation{
				add:    []float64{20},
				update: map[int]float64{1: 10},
			},
			expected: []changeDisksPostValues{
				{
					ID:  "1",
					Add: "20gb",
				},
				{
					ID:     "1",
					Resize: "1",
					Size:   "10gb",
				},
			},
		},
		{
			name: "update and remove",
			op:   diskOperation{remove: []int{1}, update: map[int]float64{0: 10}},
			expected: []changeDisksPostValues{
				{
					ID:     "1",
					Remove: "1",
				},
				{
					ID:     "1",
					Resize: "0",
					Size:   "10gb",
				},
			},
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			called := 0
			var bodies []changeDisksPostValues
			prevRequest := mockableRequest
			// request is mocked to return only body payload
			mockableRequest = func(provider *ProviderConfig, method string, path string, body interface{}) (interface{}, error) {
				called += 1
				bodies = append(bodies, body.(changeDisksPostValues))
				return []interface{}{""}, nil
			}
			defer func() {
				mockableRequest = prevRequest
			}()

			err := changeDisks(nil, "1", test.op)

			assert.Nil(t, err)
			assert.Equal(t, len(test.expected), called)
			assert.Equal(t, test.expected, bodies)
		})
	}
}
