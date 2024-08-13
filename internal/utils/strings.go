package utils

import "strings"

func JoinStrings(joinChar string, strs ...string) string {
	if len(strs) == 0 {
		return ""
	}

	var result strings.Builder

	for i, s := range strs {
		if i > 0 {
			lastChar := ""
			if result.Len() > 0 {
				lastChar = result.String()[result.Len()-1:]
			}
			nextFirstChar := ""
			if len(s) > 0 {
				nextFirstChar = s[:1]
			}

			if lastChar != joinChar && nextFirstChar != joinChar {
				result.WriteString(joinChar)
			}
		}
		result.WriteString(s)
	}

	return result.String()
}
