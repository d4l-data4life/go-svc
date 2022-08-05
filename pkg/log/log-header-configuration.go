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

var hlcInRequest *headerObfuscator = newHeaderObfuscator().
	obfuscateHeaders(defaultObfuscationHeaders).
	ignoreHeaders(defaultIgnoreHeaders)

var hlcInResponse *headerObfuscator = newHeaderObfuscator().
	obfuscateHeaders(defaultObfuscationHeaders).
	ignoreHeaders(defaultIgnoreHeaders)

var hlcOutRequest *headerObfuscator = newHeaderObfuscator().
	obfuscateHeaders(defaultObfuscationHeaders).
	ignoreHeaders(defaultIgnoreHeaders)

var hlcOutResponse *headerObfuscator = newHeaderObfuscator().
	obfuscateHeaders(defaultObfuscationHeaders).
	ignoreHeaders(defaultIgnoreHeaders)
