package utils

func DedupStrings(strs []string) []string {
	m := make(map[string]bool)
	deduped := []string{}

	for _, str := range strs {
		if !m[str] {
			m[str] = true
			deduped = append(deduped, str)
		}
	}

	return deduped
}

func HasString(slice []string, str string) (int, bool) {
	for i, s := range slice {
		if s == str {
			return i, true
		}
	}
	return -1, false
}
