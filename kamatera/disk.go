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

	var oldValues []int
	for _, v := range o.([]interface{}) {
		switch v.(type) {
		case float64:
			oldValues = append(oldValues, int(v.(float64)))
		case int:
			oldValues = append(oldValues, v.(int))
		default:
			return diskOperation{}, cannotParseDiskValuesErr{
				old: o,
				new: n,
			}
		}
	}
	var newValues []int
	for _, v := range n.([]interface{}) {
		switch v.(type) {
		case float64:
			newValues = append(newValues, int(v.(float64)))
		case int:
			newValues = append(newValues, v.(int))
		default:
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
				op.update = map[int]int{}
			}
			op.update[i] = newValues[i]
		}
	}
	return op, nil
}
