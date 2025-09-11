package log

import (
	"bufio"
	"net/http"
	"net/textproto"
	"reflect"
	"strings"
	"testing"

	"github.com/d4l-data4life/go-svc/pkg/log/testutils"
)

func Test_headerObfuscator_obfuscateHeaders(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want *headerObfuscator
	}{
		{
			name: "Canonicalization",
			args: []string{"Authorization", "authorization", "AUTHORIZATION"},
			want: &headerObfuscator{
				obfuscateHeader: map[string]bool{textproto.CanonicalMIMEHeaderKey("Authorization"): true},
				ignoreHeader:    map[string]bool{},
			},
		},
		{
			name: "Adding Set-Cookie & Cookie which is handled by default",
			args: []string{"Set-Cookie", "Cookie"},
			want: &headerObfuscator{
				map[string]bool{},
				map[string]bool{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			heob := newHeaderObfuscator()
			if got := heob.obfuscateHeaders(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("headerObfuscator.obfuscateHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_headerObfuscator_ignoreHeaders(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want *headerObfuscator
	}{
		{
			name: "Canonicalization",
			args: []string{"Authorization", "authorization", "AUTHORIZATION"},
			want: &headerObfuscator{
				obfuscateHeader: map[string]bool{},
				ignoreHeader:    map[string]bool{textproto.CanonicalMIMEHeaderKey("Authorization"): true},
			},
		},
		{
			name: "Adding Set-Cookie & Cookie which is handled by default",
			args: []string{"Set-Cookie", "Cookie"},
			want: &headerObfuscator{
				map[string]bool{},
				map[string]bool{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			heob := newHeaderObfuscator()
			if got := heob.ignoreHeaders(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("headerObfuscator.ignoreHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_headerObfuscator_Disjunct(t *testing.T) {
	type operationOrder int
	const (
		ignore_obfuscate operationOrder = 0
		obfuscate_ignore operationOrder = 1
	)

	canonicalize := textproto.CanonicalMIMEHeaderKey

	tests := []struct {
		name           string
		argsIgnore     []string
		argsObfuscate  []string
		operationOrder operationOrder
		want           *headerObfuscator
	}{
		{
			name:           "Disjunct headers in ignore and obfuscate set",
			argsIgnore:     []string{"X-Test", "X-TEST", "X-Ignore"},
			argsObfuscate:  []string{"x-test", "X-obfuscate", "X-Obfuscate"},
			operationOrder: ignore_obfuscate,
			want: &headerObfuscator{
				ignoreHeader:    map[string]bool{canonicalize("X-Test"): true, canonicalize("X-Ignore"): true},
				obfuscateHeader: map[string]bool{canonicalize("X-Obfuscate"): true},
			},
		}, {
			name:           "Disjunct headers in obfuscate and ignore set",
			argsIgnore:     []string{"X-Test", "x-ignore", "X-Ignore"},
			argsObfuscate:  []string{"x-test", "X-obfuscate", "X-Obfuscate"},
			operationOrder: obfuscate_ignore,
			want: &headerObfuscator{
				ignoreHeader:    map[string]bool{canonicalize("X-Ignore"): true},
				obfuscateHeader: map[string]bool{canonicalize("X-Test"): true, canonicalize("X-Obfuscate"): true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newHeaderObfuscator()

			if tt.operationOrder == obfuscate_ignore {
				got = got.obfuscateHeaders(tt.argsObfuscate).ignoreHeaders(tt.argsIgnore)
			} else {
				got = got.ignoreHeaders(tt.argsIgnore).obfuscateHeaders(tt.argsObfuscate)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("headerObfuscator.ignoreHeaders/obfuscateHeaders() = \n%v,\n want =\n%v", got, tt.want)
			}
		})
	}
}

const SAMPLEREQUEST = `GET localhost:8080/headers HTTP/1.1
User-Agent: curl/7.77.0
Accept: */*
Cookie: jwt=secret; nonjwt=nosecret
Authorization: bearer OmTFFtrcyqn9DBdf0vKDCQ
authorization: bearer itrytobecheekyhere
Content-Type: application/json
Set-cookie: jwt=secret; HttpOnly;
Set-cookie: jwt=thiscookieshouldbeoverridden
Set-cookie: nonjwt=nosecret
Connection: Close
X-Real-Ip: 192.168.0.8

{}
`

func Test_headerObfuscator_ProcessHeaders(t *testing.T) {
	tests := []struct {
		name string
		heob *headerObfuscator
		args *http.Request
		want http.Header
	}{{
		name: "Ignore headers with canonicalization",
		heob: newHeaderObfuscator().ignoreHeaders([]string{"X-Real-Ip", "content-type"}).obfuscateHeaders([]string{"Authorization"}),
		args: testutils.Request(
			testutils.WithHeader("x-real-ip", "test"),
			testutils.WithAppendToHeader("x-real-ip", "moretest"),
			testutils.WithHeader("CONTENT-TYPE", "test2"),
		),
		want: map[string][]string{},
	}, {
		name: "Obfuscate headers with canonicalization",
		heob: newHeaderObfuscator().ignoreHeaders([]string{"X-Real-Ip", "content-type"}).obfuscateHeaders([]string{"Authorization"}),
		args: testutils.Request(
			testutils.WithHeader("authorization", "test"),
			testutils.WithAppendToHeader("Authorization", "test2"),
		),
		want: map[string][]string{
			"Authorization": {"Obfuscated{4}", "Obfuscated{5}"}},
	}, {
		name: "Default cookie value obfuscation",
		heob: newHeaderObfuscator().ignoreHeaders([]string{"X-Real-Ip", "content-type"}).obfuscateHeaders([]string{"Authorization"}),
		args: testutils.Request(
			testutils.WithCookies(
				&http.Cookie{Name: "awskey", Value: "secret"},
				&http.Cookie{Name: "tracking", Value: "nosecret"}),
		),
		want: map[string][]string{
			"Cookie": {"awskey=Obfuscated{6}; tracking=Obfuscated{8};"}},
	}, {
		name: "Default set-cookie value obfuscation",
		heob: newHeaderObfuscator().ignoreHeaders([]string{"X-Real-Ip", "content-type"}).obfuscateHeaders([]string{"Authorization"}),
		args: testutils.Request(
			testutils.WithHeader("Set-Cookie", "key=value"),
			testutils.WithAppendToHeader("Set-Cookie", "key2=value+;"),
			testutils.WithAppendToHeader("Set-Cookie", "key3=value++; Path=/path/; Domain=example.com"),
			testutils.WithAppendToHeader("Set-Cookie", "bad"),
		),
		want: map[string][]string{
			"Set-Cookie": {
				"key=Obfuscated{5}",
				"key2=Obfuscated{6};",
				"key3=Obfuscated{7}; Path=/path/; Domain=example.com",
				"Invalid{3}",
			}},
	}, {
		name: "Mixing everything",
		heob: newHeaderObfuscator().ignoreHeaders([]string{"X-Real-Ip", "content-type"}).obfuscateHeaders([]string{"Authorization"}),
		args: func() *http.Request {
			req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(SAMPLEREQUEST)))
			if err != nil {
				panic("Could not parse HTTP request")
			}
			return req
		}(),
		want: map[string][]string{
			"User-Agent":    {"curl/7.77.0"},
			"Accept":        {"*/*"},
			"Cookie":        {"jwt=Obfuscated{6}; nonjwt=Obfuscated{8};"},
			"Authorization": {"Obfuscated{29}", "Obfuscated{25}"},
			"Set-Cookie":    {"jwt=Obfuscated{6}; HttpOnly;", "jwt=Obfuscated{28}", "nonjwt=Obfuscated{8}"},
			"Connection":    {"Close"},
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.heob.processHeaders(tt.args.Header)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("headerObfuscator.ProcessHeaders() got = \n%v\nwant = \n%v", got, tt.want)
			}
		})
	}
}
