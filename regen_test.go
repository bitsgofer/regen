package regen

import (
	"bytes"
	"regexp/syntax"
	"testing"
)

func TestGenStringWithRandSrc(t *testing.T) {
	theAnswerToEverything := func(n int64) int64 { // affects both number of repetitions && exact character used
		return 2
	}
	pattern := `#-[[:digit:]]{2,5}`
	var b bytes.Buffer
	re, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		t.Fatalf("cannot parse pattern to regex; err= %v", err)
	}

	if err := GenString(&b, re, RandSrc(theAnswerToEverything)); err != nil {
		t.Fatalf("want no error, got= %v", err)
	}
	if want, got := "#-2222", b.String(); want != got {
		t.Fatalf("want= %v, got= %v", want, got)
	}
}

func TestGenString(t *testing.T) {
	var testCases = []struct { // execute in strict order
		name    string
		pattern string
		options []Option

		isErr bool
		str   string
	}{
		{
			name:    "straightMatch",
			pattern: `noAmbigu1ty!`,
			str:     "noAmbigu1ty!",
		},
		{
			name:    "charClass",
			pattern: `#-[[:digit:]]{2,5}`, // appartment number
			str:     "#-1117",
		},
		{
			name:    "captureGroup",
			pattern: `127(\.[[:digit:]]){3}`, // IPv4
			str:     "127.8.8.6",
		},
		{
			name:    "nested",
			pattern: `00:(:[0-9a-z]{2}){5}`, // IPv6
			str:     "00::3y:zt:je:hk:1a",
		},
		{
			name:    "nestedCaptureGroup",
			pattern: `((((((((((((((((((((((((((((((x)+)))y))))))))))))))))))))))))))*`,
			str:     "xxxxxxxxxxxxxxxxxxxxxxxxxyxxxxxxxxxxxxxxxxxxxxxxxxxxyxxxxxxxxxxxxxxyxxxxxxyxxxxxxxxxxxxxxxxxxxxxxxxxxxxyxxxxxxxxxxxxxxxxxxyxxxxxxxxxxxxxxxxxxxxxxxxxxxxyxxxxxxyxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxyxxxxxxxxxxxxxxxxxxxxxxxxxxyxxxxxxxxxxxxxxxxxxxxxxxyxxxxxxxxxxxxxxxxxxxxxxxxxxxyxxxxyxxxxxxxxxxxxxxxxyxxxxxyxxxxxxxxyxxxxxxxyxxxxxxxxxxxxxxxxxxxxyxxxxxxxxxxxxxxxxxxxxxxyxxxxxxxxxxyxyxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxyxxxxxxxxxxxxxxxxxxxxxxxyxxxxxxxxxxxxxxxxxxxxxxxxxxxxyxxxxxxxxyxxxxxxxxxxxxxxxxyxxxxxxxxxxxy",
		},
		{
			name:    "useBoth^And$",
			pattern: "^hello$",
			isErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			re, err := syntax.Parse(tc.pattern, syntax.Perl)
			if err != nil {
				t.Fatalf("cannot parse pattern to regex; err= %v", err)
			}
			err = GenString(&b, re, tc.options...)
			switch {
			case tc.isErr && err == nil:
				t.Fatalf("want error, got none")
			case tc.isErr && err != nil:
				return
			case !tc.isErr && err != nil:
				t.Fatalf("want no error, got= %v", err)
			}

			if want, got := tc.str, b.String(); want != got {
				t.Fatalf("want= %v, got= %v", want, got)
			}
		})
	}
}
