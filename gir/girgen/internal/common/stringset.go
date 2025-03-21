package common

// StringSet joins the given slices of strings into a map with the keys as the
// values of each of the given slices.
func StringSet(strs ...[]string) map[string]struct{} {
	var length int
	for _, str := range strs {
		length += len(str)
	}

	set := make(map[string]struct{}, length)

	for _, str := range strs {
		for _, s := range str {
			set[s] = struct{}{}
		}
	}

	return set
}
