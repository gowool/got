package internal

// Seq creates a sequence of integers from args.
//
// Examples:
//
//	3 => 1, 2, 3
//	1 2 4 => 1, 3
//	-3 => -1, -2, -3
//	1 4 => 1, 2, 3, 4
//	1 -2 => 1, 0, -1, -2
func Seq(args ...int) []int {
	if len(args) < 1 || len(args) > 3 {
		// invalid number of arguments to Seq
		return nil
	}

	inc := 1
	var last int
	first := args[0]

	if len(args) == 1 {
		last = first
		if last == 0 {
			return nil
		} else if last > 0 {
			first = 1
		} else {
			first = -1
			inc = -1
		}
	} else if len(args) == 2 {
		last = args[1]
		if last < first {
			inc = -1
		}
	} else {
		inc = args[1]
		last = args[2]
		if inc == 0 {
			// 'increment' must not be 0
			return nil
		}
		if first < last && inc < 0 {
			// 'increment' must be > 0
			return nil
		}
		if first > last && inc > 0 {
			// 'increment' must be < 0
			return nil
		}
	}

	// sanity check
	if last < -100000 {
		// size of result exceeds limit
		return nil
	}
	size := ((last - first) / inc) + 1

	// sanity check
	if size <= 0 || size > 2000 {
		// size of result exceeds limit
		return nil
	}

	seq := make([]int, size)
	val := first
	for i := 0; ; i++ {
		seq[i] = val
		val += inc
		if (inc < 0 && val < last) || (inc > 0 && val > last) {
			break
		}
	}

	return seq
}
