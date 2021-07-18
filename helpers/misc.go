package helpers

// MakeUnique remove dup from a slice
func MakeUnique(names []string) []string {
	// if the name occurs the flag changes to true.
	// hence all name that has already occurred will be true.
	flag := make(map[string]bool)
	var uniqueNames []string
	for _, name := range names {
		if flag[name] == false {
			flag[name] = true
			uniqueNames = append(uniqueNames, name)
		}
	}
	// unique names collected
	return uniqueNames
}
