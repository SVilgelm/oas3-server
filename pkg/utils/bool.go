package utils

// DoubleMapBool is a map of map of string
type DoubleMapBool map[string]map[string]bool

// Set set a value of a double map
func (d *DoubleMapBool) Set(first, second string, value bool) {
	dest := *d
	if _, ok := dest[first]; !ok {
		dest[first] = make(map[string]bool)
	}
	dest[first][second] = value
}
