package main

import (
	"strconv"
	"strings"
)

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func replaceVariable(variable string, toReplace *string, source *string) string {
	quoted := strconv.Quote(*toReplace) // escape all escape-chars
	quoted = quoted[1 : len(quoted)-1]  // remove the quotes the previous function added

	return strings.ReplaceAll(*source, variable, quoted)
}
