package ternary

// If implements a conditional expression that returns one of two values based on a condition
// If 函数实现了一个条件表达式，根据条件返回两个值中的一个
//
// Parameters 参数:
//   - condition: boolean value that determines which value to return
//     布尔值，用于决定返回哪个值
//   - trueVal: value to return if condition is true
//     当 condition 为 true 时返回的值
//   - falseVal: value to return if condition is false
//     当 condition 为 false 时返回的值
//
// Returns 返回值:
//   - T: either trueVal or falseVal depending on condition
//     根据条件返回 trueVal 或 falseVal
func If[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}
