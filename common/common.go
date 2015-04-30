package common

import ()

func IsExistInArray(element string, list []string) bool {
	for _, item := range list {
		if item == element {
			return true
		}
	}
	return false
}
