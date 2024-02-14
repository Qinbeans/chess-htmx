package utils

// abs returns the absolute value of x for integers
func Abs(x int) int {
	return x &^ (x >> 31)
}
