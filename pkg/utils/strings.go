package utils

// Contains validates that ss contains s
func Contains(ss []string, s string) bool {
	for _, v := range ss {
		if s == v {
			return true
		}
	}
	return false
}

// DoubleMapString is a map of map of string
type DoubleMapString map[string]map[string]string

// Set set a value of a double map
func (d *DoubleMapString) Set(first, second, value string) {
	dest := *d
	if _, ok := dest[first]; !ok {
		dest[first] = make(map[string]string)
	}
	dest[first][second] = value
}
