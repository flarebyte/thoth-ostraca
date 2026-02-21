package config

import "strings"

const CurrentConfigVersion = "1"

var SupportedConfigVersions = []string{CurrentConfigVersion}

func IsSupportedConfigVersion(v string) bool {
	for _, s := range SupportedConfigVersions {
		if v == s {
			return true
		}
	}
	return false
}

func SupportedConfigVersionsCSV() string {
	return strings.Join(SupportedConfigVersions, ", ")
}
