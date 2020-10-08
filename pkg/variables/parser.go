package variable

import "strings"

func ReplaceVariables(vars VariableMap, s string) string {
	if vars == nil {
		return s
	}

	for k, v := range vars {
		s = strings.Replace(s, k, v.Value, 1)
	}

	return s
}