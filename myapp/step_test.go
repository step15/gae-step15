package step

import (
	"reflect"
	"testing"
)

func TestUrlRE(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{{`http://step15-krispop.appspot.com/`, true},
		{`http://step15-krispop.appspot.com`, true},
		{`http://step15-<img>-foo.appspot.com`, false},
		{`http://foo.appspot.com`, true},
	}
	for _, test := range tests {
		if got := appspotMatchRe.MatchString(test.in); got != test.want {
			t.Errorf("match url: %v got %v want %v", test.in, got, test.want)
		}
	}
}

func TestParsePeers(t *testing.T) {
	in := `http://step15-krispop.appspot.com/	F	T	F	F	F
http://regal-sun-100211.appspot.com	T	F	F	F	F`
	got := parsePeers(in)
	want := map[string][]string{
		"show":    []string{"http://step15-krispop.appspot.com"},
		"convert": []string{"http://regal-sun-100211.appspot.com"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parsePeers(%v) = %v, want %v", in, got, want)
	}
}
