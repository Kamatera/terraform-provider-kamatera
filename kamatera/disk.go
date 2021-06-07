package kamatera

import (
	"fmt"
)

type cannotParseDiskValuesErr struct {
	old interface{}
	new interface{}
}

func (c cannotParseDiskValuesErr) Error() string {
	return fmt.Sprintf("cannot parse disk values: o: %v, n: %v", c.old, c.new)
}

func calDiskChangeOperation(o, n interface{}) (diskOperation, error) {
	op := diskOperation{}

	var oldValues []float64
	for _, v := range o.([]interface{}) {
		val, ok := v.(float64)
		if ok {
			oldValues = append(oldValues, val)
		} else {
			return diskOperation{}, cannotParseDiskValuesErr{
				old: o,
				new: n,
			}
		}
	}
	var newValues []float64
	for _, v := range n.([]interface{}) {
		val, ok := v.(float64)
		if ok {
			newValues = append(newValues, val)
		} else {
			return diskOperation{}, cannotParseDiskValuesErr{
				old: o,
				new: n,
			}
		}
	}
	if len(oldValues) == 0 || len(newValues) == 0 {
		return diskOperation{}, cannotParseDiskValuesErr{
			old: o,
			new: n,
		}
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
