package utils

// abs returns the absolute value of x for integers
func Abs(x int) int {
	return x &^ (x >> 31)
}

// Min returns the minimum of x and y
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Max returns the maximum of x and y
func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
