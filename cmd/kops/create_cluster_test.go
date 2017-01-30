package main

import (
	"testing"
)

func TestParseCloudLabels(t *testing.T) {
	expect := map[string]string{"foo":"bar", "fib":"baz"}
	checkParse(t, make([]string,0), map[string]string{}, false)
	checkParse(t, []string{"foo=bar","fib=baz"}, expect, false)
	checkParse(t, []string{`foo=bar`,`"fib"="baz"`}, expect, false)
	checkParse(t, []string{`"f,o\""o"=bar`,`"fi\b"="baz"`},
		map[string]string{`f,o\"o`:"bar", `fi\b`:"baz"}, false)
	checkParse(t, []string{`fo"o=bar`,`fib=baz`}, expect, true)
}

func checkParse(t *testing.T, s []string, expect map[string]string, shouldErr bool) {
	m, err := parseCloudLabels(s)
	if err != nil {
		if shouldErr {
			return
		} else {
			t.Errorf(err.Error())
		}
	}

	for k, v := range expect {
		if m[k] != v {
			t.Errorf("Expected: %v, Got: %v", expect, m)
		}
	}
}
