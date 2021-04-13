package strings

// IndexOf returns index position in slice from given string
// If value is -1, the string does not found.
func IndexOf(slice []string, s string) int {
	for i, v := range slice {
		if v == s {
			return i
		}
	}

	return -1
}

// Include returns true or false if given string is in slice.
func Include(slice []string, s string) bool {
	return IndexOf(slice, s) >= 0
}

// UniqueAppend appends a string if not exist in the slice.
func UniqueAppend(slice []string, s ...string) []string {
	for i := range s {
		if IndexOf(slice, s[i]) != -1 {
			continue
		}

		slice = append(slice, s[i])
	}

	return slice
}

// Reverse reverses a string slice.
func Reverse(slice []string) []string {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}

	return slice
}

// Copy returns a copy of the slice.
func Copy(slice []string) []string {
	dst := make([]string, len(slice))
	copy(dst, slice)

	return dst
}

// Equal checks if the slices are equal.
func Equal(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}

	return true
}
