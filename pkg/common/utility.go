package common

import (
	"math/rand"
)

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

func GetRandInt(min, max, exception int) int {
	rnd := rand.Intn(max - min) + min
	for rnd == exception {
		rnd = GetRandInt(min, max, exception)
	}
	return rnd
}