package model

import "strings"

// CanonicalIdentifier keeps station and sequence keys stable across receivers
// that differ only in casing or accidental surrounding whitespace. Lowercasing
// makes the receiver and sequence identity case-insensitive so that a retransmit
// with "North" / "north" or "  seq-1  " / "seq-1" resolves to the same record.
func CanonicalIdentifier(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func ValidIdentifier(value string) bool {
	if value == "" || len(value) > 96 {
		return false
	}
	for _, runeValue := range value {
		if runeValue >= 'a' && runeValue <= 'z' {
			continue
		}
		if runeValue >= '0' && runeValue <= '9' {
			continue
		}
		if runeValue == '-' || runeValue == '_' || runeValue == '.' {
			continue
		}
		return false
	}
	return true
}
