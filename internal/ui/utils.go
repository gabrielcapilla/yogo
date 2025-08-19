package ui

func truncate(s string, length int) string {
	if length <= 3 {
		return "..."
	}
	if len(s) > length {
		return s[:length-3] + "..."
	}
	return s
}
