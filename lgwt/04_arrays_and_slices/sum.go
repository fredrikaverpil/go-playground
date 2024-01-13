package arrays_and_slices

func SumArray(numbers [5]int) int {
	sum := 0
	for _, n := range numbers {
		sum += n
	}
	return sum
}

func SumSlice(numbers []int) int {
	sum := 0
	for _, n := range numbers {
		sum += n
	}
	return sum
}

func SumAll(numbersToSum ...[]int) []int {
	var sums []int

	for _, numbers := range numbersToSum {
		sums = append(sums, SumSlice(numbers))
	}

	return sums
}

func SumAllTails(numbersToSum ...[]int) []int {
	var sums []int

	for _, numbers := range numbersToSum {
		if len(numbers) <= 1 {
			sums = append(sums, 0)
		} else {
			tail := numbers[1:]
			sums = append(sums, SumSlice(tail))

		}
	}

	return sums
}
