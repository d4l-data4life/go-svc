package log

import (
	"regexp"
	"strings"
)

var MOBILE_REGEX = regexp.MustCompile(`\d{1,3}\d{7,}`)

// MobileObfuscator obfuscates mobile phone numbers in log messages.
// `WithPlusPrefix` indicates if the obfuscator should only recognize sequences as mobile numbers that start with a '+'
type MobileObfuscator struct {
	EventType EventType
	ReqMethod string
	Replace   *regexp.Regexp
	Field     Field
}

func (mo MobileObfuscator) GetEventType() EventType {
	return mo.EventType
}

func (mo MobileObfuscator) GetReqMethod() string {
	return mo.ReqMethod
}

func (mo MobileObfuscator) Obfuscate(log interface{}) interface{} {
	switch l := log.(type) {
	case inRequestLog:
		switch mo.Field {
		case URL:
			l.ReqURL = ObfuscateMobile(l.ReqURL, mo.Replace)
			return l
		case Body:
			l.ReqBody = ObfuscateMobile(l.ReqBody, mo.Replace)
			return l
		}
		return l
	case outRequestLog:
		switch mo.Field {
		case URL:
			l.ReqURL = ObfuscateMobile(l.ReqURL, mo.Replace)
			return l
		case Body:
			l.ReqBody = ObfuscateMobile(l.ReqBody, mo.Replace)
			return l
		}
		return l
	case outResponseLog:
		switch mo.Field {
		case URL:
			l.ReqURL = ObfuscateMobile(l.ReqURL, mo.Replace)
			return l
		case Body:
			l.ResponseBody = ObfuscateMobile(l.ResponseBody, mo.Replace)
			return l
		}
		return l
	}

	return log
}

// Obfuscates a mobile number to be GDPR compliant while trying to keep the logs debuggable.
// Obfuscates the middle part of the mobile number.
// See https://gesundheitscloud.atlassian.net/wiki/spaces/PHDP/pages/2407038985/Logging+in+a+GDPR+compliant+way#Mobile-number-filter
func ObfuscateMobile(raw string, replace *regexp.Regexp) string {

	for _, replaceMatch := range replace.FindAllStringSubmatch(raw, 1) {
		matches := MOBILE_REGEX.FindAllStringSubmatch(replaceMatch[0], -1)

		for _, match := range matches {
			obf := processNumber(match[0])

			raw = strings.Replace(raw, match[0], obf, 1)
		}
	}

	return raw
}

func processNumber(number string) string {
	obfNumber := []rune(number)

	len := len(number)
	obfCount := len / 2
	obfCountEven := obfCount%2 == 0
	charCountEven := len%2 == 0

	// Check if the range of obfuscation characters fits symmetrically into the mobile number
	// and if not, emphasize the right side.
	// Symmetry is possible, when both obfCount and charCount are even or uneven.
	startPos := (len - obfCount) / 2
	if obfCountEven != charCountEven {
		startPos += 1
	}

	for i := 0; i < obfCount; i += 1 {
		obfNumber[startPos+i] = OBF_CHAR
	}

	return string(obfNumber)
}
