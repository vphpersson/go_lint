package clean

import "slices"

type option func(*int)

// Appending to a copy is the fix, and must not be flagged.
func cloned(options ...option) []option {
	return append(slices.Clone(options), nil)
}

// A fresh slice the caller never saw.
func freshSlice(options ...option) []option {
	copied := make([]option, 0, len(options)+1)
	copied = append(copied, options...)
	copied = append(copied, nil)
	return copied
}

// A literal is not the caller's array either.
func literalFirst(options ...option) []option {
	return append([]option{nil}, options...)
}

// An ordinary slice parameter is the caller's to reason about, not this analyzer's business.
func ordinarySlice(options []option) []option {
	return append(options, nil)
}

// A local variable shadowing the parameter name is a different object.
func shadowed(options ...option) []option {
	{
		options := make([]option, 0)
		options = append(options, nil)
		_ = options
	}
	return options
}

// append with nothing to add writes nothing.
func noAddition(options ...option) []option {
	return append(options)
}

// A local append is not a parameter append.
func localOnly() []option {
	var options []option
	return append(options, nil)
}
