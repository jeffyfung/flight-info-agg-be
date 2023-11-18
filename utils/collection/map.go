package collection

func Map[T, U any](array []T, f func(T) U) []U {
	outputArr := make([]U, len(array))
	for i := range array {
		outputArr[i] = f(array[i])
	}
	return outputArr
}