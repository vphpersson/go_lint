package flagged

type option func(*int)

func direct(options ...option) []option {
	return append(options, nil) // want `append to variadic parameter "options" writes into the caller's array`
}

func assignedBack(options ...option) []option {
	options = append(options, nil) // want `append to variadic parameter "options"`
	return options
}

func throughSlice(options ...option) []option {
	return append(options[:0], nil) // want `append to variadic parameter "options"`
}

func spread(options ...option) []option {
	return append(options, options...) // want `append to variadic parameter "options"`
}

func insideClosure(options ...option) func() []option {
	return func() []option {
		return append(options, nil) // want `append to variadic parameter "options"`
	}
}

func closureParameter() func(...option) []option {
	return func(options ...option) []option {
		return append(options, nil) // want `append to variadic parameter "options"`
	}
}

func alsoStrings(names ...string) []string {
	return append(names, "x") // want `append to variadic parameter "names"`
}
