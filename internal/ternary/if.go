package ternary

// If 在 Go 中提供简洁的三目表达式语义。
func If[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}
