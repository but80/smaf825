package util

// GCD は、2つの整数の最大公約数を求めます。
func GCD(a, b int) int {
	for a != b {
		if a < b {
			b -= a
		} else {
			a -= b
		}
	}
	return a
}
