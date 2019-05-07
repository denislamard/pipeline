package pipeline

import "strings"

func parseEnvironment(name string) bool {
	switch strings.ToUpper(name) {
	case "DEBUG":
		return true
	case "INTEGRATION":
		return true
	case "PRODUCTION":
		return true
	default:
		return false
	}
}

