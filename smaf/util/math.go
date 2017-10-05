package util

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
