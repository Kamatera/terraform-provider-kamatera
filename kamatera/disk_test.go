package kamatera

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_calDiskChangeOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		o           []interface{}
		n           []interface{}
		expected    diskOperation
		expectedErr error
	}{
		{
			name:     "update only",
			o:        []interface{}{123.0, 456.0},
			n:        []interface{}{123.0, 457.0},
			expected: diskOperation{update: map[int]float64{1: 457}},
		},
		{
			name:     "remove only",
			o:        []interface{}{123.0, 456.0},
			n:        []interface{}{123.0},
			expected: diskOperation{remove: []int{1}},
		},
		{
			name: "add only",
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
		{
			name: "cannot parse old values",
			o: []interface{}{"a", 123.0},
			n: []interface{}{123.0},
			expected: diskOperation{},
			expectedErr: cannotParseDiskValuesErr{old: []interface{}{"a", 123.0},new: []interface{}{123.0}},
		},
		{
			name: "has 2 same size disks and resize 1",
			o: []interface{}{123.0, 123.0},
			n: []interface{}{123.0, 456.0},
			expected: diskOperation{update: map[int]float64{1: 456.0}},
		},
		{
			name: "update 4 disks",
			o: []interface{}{1.0, 2.0, 3.0, 4.0},
			n: []interface{}{2.0, 3.0, 4.0, 5.0},
			expected: diskOperation{update: map[int]float64{
				0: 2,
				1: 3,
				2: 4,
				3: 5,
			}},
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
