package utils

func SafeNewSliceGenerator(begin, end, step int) []int {
	var sequence []int
	if step == 0 {
		step = 1
	}
	count := 0
	if (end > begin && step > 0) || (end < begin && step < 0) {
		count = (end-step-begin)/step + 1
	}

	sequence = make([]int, count)
	for i := 0; i < count; i, begin = i+1, begin+step {
		sequence[i] = begin
	}
	return sequence
}
