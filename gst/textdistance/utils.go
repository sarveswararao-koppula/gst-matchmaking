package textdistance

import "strings"

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clearStr(str string) string {
	str = strings.ToLower(str)
	return strings.Join(strings.Fields(str), " ")
}

