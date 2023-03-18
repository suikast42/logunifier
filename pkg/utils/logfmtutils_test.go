package utils

import (
	"reflect"
	"strings"
	"testing"
)

func TestIsKey(t *testing.T) {

	var word = "a=1 b=1 d="
	word, _isKey := isKey("a", word)
	if !_isKey {
		t.Errorf("a should be key. Restword %s", word)
	}
	if !strings.EqualFold("b=1 d=", word) {
		t.Errorf("Expected Restword [%s] but was [%s]", "b=1 d=", word)
	}

	word, _isKey = isKey("b", word)
	if !_isKey {
		t.Errorf("b should be key. Restword %s", word)
	}

	if !strings.EqualFold("d=", word) {
		t.Errorf("Expected Restword [%s] but was [%s]", "d=", word)
	}

	word, _isKey = isKey("d", word)
	if !_isKey {
		t.Errorf("d should be key. Restword %s", word)
	}
	if !strings.EqualFold("", word) {
		t.Errorf("Expected Restword [%s] but was [%s]", "", word)
	}

	word, _isKey = isKey("d", word)
	if _isKey {
		t.Errorf("d should not be a key. Restword %s", word)
	}
	if !strings.EqualFold("", word) {
		t.Errorf("Expected Restword [%s] but was [%s]", "", word)
	}
}
func TestValidKvs(t *testing.T) {
	tests := []struct {
		pos  int
		data string
		want map[string]string
	}{
		{
			pos:  1,
			data: "a=1",
			want: map[string]string{
				"a": "1",
			},
		},

		{
			pos:  2,
			data: "a=1 b=2",
			want: map[string]string{
				"a": "1",
				"b": "2",
			},
		},

		{
			pos:  3,
			data: "a=1 b=1 d=",
			want: map[string]string{
				"a": "1",
				"b": "1",
				"d": "",
			},
		},
		{
			pos:  4,
			data: "a=1 b=1 d=\"\"",
			want: map[string]string{
				"a": "1",
				"b": "1",
				"d": "",
			},
		},
		{
			pos:  5,
			data: "a=1 b=1 multiline=\"line1\nline2\"",
			want: map[string]string{
				"a":         "1",
				"b":         "1",
				"multiline": "line1\nline2",
			},
		},

		{
			pos:  6,
			data: "multiline=\"line1\nline2\"",
			want: map[string]string{
				"multiline": "line1\nline2",
			},
		},

		{
			pos:  7,
			data: "a= b= c=2",
			want: map[string]string{
				"a": "",
				"b": "",
				"c": "2",
			},
		},
		{
			pos:  8,
			data: "a@1=2 b= c=2",
			want: map[string]string{
				"a@1": "2",
				"b":   "",
				"c":   "2",
			},
		},
	}

	for _, test := range tests {

		parsed, err := DecodeLogFmt(test.data)

		if err != nil {
			t.Errorf("In pos: %d. Expected no error but got %+v. For input string %s. Parsed is %+v", test.pos, err, test.data, parsed)
		}

		if len(parsed) != len(test.want) {
			t.Errorf("In pos: %d. Expect %d keys but got %d", test.pos, len(test.want), len(parsed))
		}
		if !reflect.DeepEqual(parsed, test.want) {
			t.Errorf("\npos:%d \nin: %q\nwant: %+v\ngot:  %+v", test.pos, test.data, test.want, parsed)
		}

	}
}

func TestInValidKvs(t *testing.T) {
	tests := []struct {
		pos  int
		data string
		want map[string]string
	}{
		{
			pos:  1,
			data: "you got it a=1 b= ",
			want: map[string]string{
				"a":                    "1",
				"b":                    "",
				string(LogfmtKeyTrash): "you got it",
			},
		},

		{
			pos:  2,
			data: "a=1 you got it b= ",
			want: map[string]string{
				"a":                    "1",
				"b":                    "",
				string(LogfmtKeyTrash): "you got it",
			},
		},

		{
			pos:  3,
			data: "a=1 b= you got it",
			want: map[string]string{
				"a":                    "1",
				"b":                    "",
				string(LogfmtKeyTrash): "you got it",
			},
		},

		{
			pos:  4,
			data: "ts msg level is info msg=\"the only valid stuff here\" spanID msg user not valid msg=\"is 42\"",
			want: map[string]string{
				"msg":                  "the only valid stuff here is 42",
				string(LogfmtKeyTrash): "ts msg level is info spanID msg user not valid",
			},
		},

		{
			pos:  5,
			data: "The only message here is gabare@localhost",
			want: map[string]string{
				string(LogfmtKeyMessage): "The only message here is gabare@localhost",
			},
		},
	}

	for _, test := range tests {

		parsed, err := DecodeLogFmt(test.data)

		if err == nil {
			t.Errorf("In pos: %d. Expect a arror but got nothing", test.pos)
		}

		if len(parsed) != len(test.want) {
			t.Errorf("In pos: %d. Expect %d keys but got %d", test.pos, len(test.want), len(parsed))
		}
		if !reflect.DeepEqual(parsed, test.want) {
			t.Errorf("\npos:%d \nin: %q\nwant: %+v\ngot:  %+v", test.pos, test.data, test.want, parsed)
		}

	}
}
