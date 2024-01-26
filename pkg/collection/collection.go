package collection

func Map[T, U any](array []T, f func(T) U) []U {
	outputArr := make([]U, len(array))
	for i := range array {
		outputArr[i] = f(array[i])
	}
	return outputArr
}

func RemoveListDuplicates[T comparable](inputList []T) []T {
	uniqueMap := make(map[T]struct{}) // Using an empty struct as a placeholder value

	// Iterate over the input list and add each element to the map
	for _, item := range inputList {
		uniqueMap[item] = struct{}{}
	}

	// Extract the unique keys from the map into a new slice
	uniqueList := make([]T, 0, len(uniqueMap))
	for key := range uniqueMap {
		uniqueList = append(uniqueList, key)
	}

	return uniqueList
}
