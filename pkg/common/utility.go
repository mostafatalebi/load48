package common

func ExistsStrInArray(str string, arr []string) bool {
	if arr != nil && len(arr) > 0 {
		for _, v := range arr {
			if v == str {
				return true
			}
		}
	}
	return false
}

