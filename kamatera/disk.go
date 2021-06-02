package kamatera

import (
	"fmt"
)

func calDiskChangeOperation(o, n interface{}) (diskOperation, error) {
	op := diskOperation{}

	var oldValues []float64
	for _, v := range o.([]interface{}) {
		oldValues = append(oldValues, v.(float64))
	}
	var newValues []float64
	for _, v := range n.([]interface{}) {
		newValues = append(newValues, v.(float64))
	}
	if len(oldValues) == 0 || len(newValues) == 0 {
		return op, fmt.Errorf("can not parse disk value, old: %v, new: %v", o, n)
	}

	if len(oldValues) > len(newValues) {
		oldLen := len(oldValues)
		newLen := len(newValues)
		op.remove = []int{}

		i := newLen
		for i < oldLen {
			op.remove = append(op.remove, i)
			i += 1
		}
	} else if len(oldValues) < len(newValues) {
		oldLen := len(oldValues)
		op.add = newValues[oldLen:]
	}

	// Update
	for i, v := range oldValues {
		if i == len(newValues) {
			break
		}

		if v != newValues[i] {
			if op.update == nil {
				op.update = map[int]float64{}
			}
			op.update[i] = newValues[i]
		}
	}
	return op, nil
}
