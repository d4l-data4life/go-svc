package bievents

import "strings"

type EmailType string

const (
	Internal EmailType = "internal"
	External EmailType = "external"
	Invalid  EmailType = "invalid"
)

// Create a map for internal emails for better search time complexity.
// Using empty struct here has the advantage that it doesn't require any
// additional space and Go's internal map type is optimized for that kind of values.
var internalEmails map[string]struct{} = map[string]struct{}{
	"data4life.care":      {},
	"gesundheitscloud.de": {},
	"qamadness.com":       {},
	"wearehackerone.com":  {},
	"ghostinspector.com":  {},
}

func GetEmailType(email string) (EmailType, error) {
	ed := strings.Split(email, "@")

	if len(ed) != 2 {
		return Invalid, ErrInvalidEmail
	}

	if _, ok := internalEmails[ed[1]]; ok {
		return Internal, nil
	}

	return External, nil
}

func GetEmailTypeNoError(email string) EmailType {
	t, _ := GetEmailType(email)
	return t
}
