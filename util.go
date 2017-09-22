package gold

func contains(value string, xs []string) bool {
	for _, x := range xs {
		if x == value {
			return true
		}
	}
	return false
}

func empty(xs []string) bool {
	for _, x := range xs {
		if x != "" {
			return false
		}
	}
	return true
}
