package log_test

func isMapEqual(want map[string]string, have map[string]string) bool {
	for key, value := range want {
		if value != have[key] {
			return false
		}
	}
	return true
}
