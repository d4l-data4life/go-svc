package log

import (
	"regexp"
	"strings"
)

const (
	Dot            = "."
	ObfuscatedHalf = "xxx.xxx"
)

var ipREGEX = regexp.MustCompile(`\b(25[0-5]|2[0-4]\d|[01]?\d\d?)(\.(25[0-5]|2[0-4]\d|[01]?\d\d?)){3}\b`)

// IPObfuscator obfuscates ip addresses in log messages.
type IPObfuscator struct {
	EventType EventType
	ReqMethod string
}

func (ipo IPObfuscator) GetEventType() EventType {
	return ipo.EventType
}

func (ipo IPObfuscator) GetReqMethod() string {
	return ipo.ReqMethod
}

func (ipo IPObfuscator) Obfuscate(log interface{}) interface{} {
	switch l := log.(type) {
	case inRequestLog:
		l.RealIP = ObfuscateIP(l.RealIP)
		return l
	default:
		return log
	}
}

// Obfuscates an ip address to be GDPR compliant while trying to keep the logs debuggable.
// Obfuscates the last two blocks of the ip address.
// See https://gesundheitscloud.atlassian.net/wiki/spaces/PHDP/pages/2407038985/Logging+in+a+GDPR+compliant+way#IP-filter
func ObfuscateIP(raw string) string {
	matches := ipREGEX.FindAllStringSubmatch(raw, -1)

	for _, match := range matches {
		raw = strings.Replace(raw, match[0], processIP(match[0]), 1)
	}

	return raw
}

func processIP(address string) string {
	firstDot := strings.Index(address, Dot) + 1
	secondDot := firstDot + strings.Index(address[firstDot:], Dot) + 1

	return address[:secondDot] + ObfuscatedHalf
}
