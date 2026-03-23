// File Guide for dev/ai agents:
// Purpose: Centralize the supported thoth config-version policy in one tiny place.
// Responsibilities:
// - Declare the current config version constant.
// - Expose helpers that check support and render supported versions for errors.
// - Keep version policy isolated from parsing logic.
// Architecture notes:
// - This file is intentionally small so version changes stay localized and easy to audit.
// - The supported-version list is a slice on purpose, even with one value, to keep future multi-version support straightforward.
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
