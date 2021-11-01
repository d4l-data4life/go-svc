package log

import (
	"regexp"
	"strings"
)

const OBF_CHAR = 'x'

var EMAIL_REGEX = regexp.MustCompile("([A-Z0-9a-z\\.!#$%&'*+\\-/=?^_`{|}~]+)@([A-Za-z0-9\\.\\-]+)\\.([A-Za-z]{2,64})")

// MailObfuscator obfuscates emails in log messages
type MailObfuscator struct {
	EventType EventType
	ReqMethod string
}

func (mo MailObfuscator) GetEventType() EventType {
	return mo.EventType
}

func (mo MailObfuscator) GetReqMethod() string {
	return mo.ReqMethod
}

func (mo MailObfuscator) Obfuscate(log interface{}) interface{} {
	switch l := log.(type) {
	case inRequestLog:
		l.ReqBody = ObfuscateEmail(l.ReqBody)
		l.ReqForm = ObfuscateEmail(l.ReqForm)
		return l
	case inResponseLog:
		l.ResponseBody = ObfuscateEmail(l.ResponseBody)
		return l
	case sqlLogEntry:
		l.PgxData = ObfuscateEmail(l.PgxData)
		return l
	}

	return log
}

// Obfuscates an email to be GDPR compliant while trying to keep the logs debuggable.
// Obfuscates the right half of the local part of the email and the middle part of the domain part.
// See https://gesundheitscloud.atlassian.net/wiki/spaces/PHDP/pages/2407038985/Logging+in+a+GDPR+compliant+way#Email-filter
func ObfuscateEmail(raw string) string {
	matches := EMAIL_REGEX.FindAllStringSubmatch(raw, -1)

	for _, match := range matches {
		local := match[1]
		domain := match[2]
		tld := match[3]

		obfLocal := processLocal(local)
		obfDomain := processDomain(domain)
		obfMail := obfLocal + "@" + obfDomain + "." + tld

		raw = strings.Replace(raw, match[0], obfMail, 1)
	}

	return raw
}

func processLocal(local string) string {
	obfLocal := []rune(local)

	replaceCount := len(local) / 2
	obfStartIdx := len(local) - replaceCount // obfuscate the right half
	for i := 0; i < replaceCount; i += 1 {
		obfLocal[obfStartIdx+i] = OBF_CHAR
	}

	return string(obfLocal)
}

func processDomain(domain string) string {
	obfDomain := []rune(domain)

	len := len(domain)
	obfCount := len / 2
	obfCountEven := obfCount%2 == 0
	charCountEven := len%2 == 0

	// Check if the range of OBF_CHARs fits symmetrically into the domain
	// and if not, emphasize the right side.
	// Symmetry is possible, when both obfCount and charCount are even or uneven.
	startPos := (len - obfCount) / 2
	if obfCountEven != charCountEven {
		startPos += 1
	}

	for i := 0; i < obfCount; i += 1 {
		obfDomain[startPos+i] = OBF_CHAR
	}

	return string(obfDomain)
}
