package log

var defaultObfuscationHeaders = []string{"Authorization", "WWW-Authenticate"}
var defaultIgnoreHeaders = []string{
	"X-Real-Ip",
	"X-Forwarded-For",
	"X-Scheme",
	"X-Request-Id",
	"Connection",
	"Content-Encoding",
	"Content-Type",
	"Content-Length",
	"Accept-Encoding",
	"Accept-Language",
	"Accept",
	"Date",
	"trace-id",
}

var hlcInRequest = newHeaderObfuscator().
	obfuscateHeaders(defaultObfuscationHeaders).
	ignoreHeaders(defaultIgnoreHeaders)

var hlcInResponse = newHeaderObfuscator().
	obfuscateHeaders(defaultObfuscationHeaders).
	ignoreHeaders(defaultIgnoreHeaders)

var hlcOutRequest = newHeaderObfuscator().
	obfuscateHeaders(defaultObfuscationHeaders).
	ignoreHeaders(defaultIgnoreHeaders)

var hlcOutResponse = newHeaderObfuscator().
	obfuscateHeaders(defaultObfuscationHeaders).
	ignoreHeaders(defaultIgnoreHeaders)
