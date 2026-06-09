package mathx

// 두 수의 최대공약수 (외부 공개)
func GCD(a, b int) int {
	a, b = abs(a), abs(b)
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// 비공개 헬퍼
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
