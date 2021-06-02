package kamatera

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_calDiskChangeOperation(t *testing.T) {
	tests := []struct {
		name        string
		o           []interface{}
		n           []interface{}
		expected    diskOperation
		expectedErr error
	}{
		{
			name:     "only update",
			o:        []interface{}{123.0, 456.0},
			n:        []interface{}{123.0, 457.0},
			expected: diskOperation{update: map[int]float64{1: 457}},
		},
		{
			name:     "only remove",
			o:        []interface{}{123.0, 456.0},
			n:        []interface{}{123.0},
			expected: diskOperation{remove: []int{1}},
		},
		{
			name: "only add",
			o: []interface{}{123.0},
			n: []interface{}{123.0, 456.0},
			expected: diskOperation{add: []float64{456}},
		},
		{
			name: "update and add",
			o: []interface{}{123.0},
			n: []interface{}{124.0, 456.0},
			expected: diskOperation{add: []float64{456}, update: map[int]float64{0: 124}},
		},
		{
			name: "update and remove",
			o: []interface{}{123.0, 456.0},
			n: []interface{}{124.0},
			expected: diskOperation{remove: []int{1}, update: map[int]float64{0: 124}},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			actual, actualErr := calDiskChangeOperation(test.o, test.n)
			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedErr, actualErr)
		})
	}
}
