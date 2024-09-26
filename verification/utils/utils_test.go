package utils

import (
	"reflect"
	"testing"
)

func TestPrettyPrintJson(t *testing.T) {

	var input = map[string]interface{}{"name": "Aaron", "age": 23, "visits": map[string]int{"2019": 2, "2020": 3, "2021": 0}}
	var expected = `{
  "age": 23,
  "name": "Aaron",
  "visits": {
    "2019": 2,
    "2020": 3,
    "2021": 0
  }
}`
	got := PrettyPrintJson(input)
	if got != expected {
		t.Errorf("expected:\n%s\ngot\n%s", expected, got)
	}
}

func TestSplitString(t *testing.T) {
	input := [8]string{
		"",
		" ",
		"default",
		"name,space",
		"one,two,",
		"/sub",
		"pre/sub",
		"pre/"}

	expect := [][]string{
		{""},
		{""},
		{"default"},
		{"name", "space"},
		{"one", "two", ""},
		{"", "sub"},
		{"pre", "sub"},
		{"pre", ""}}

	splited := []bool{
		false,
		false,
		false,
		true,
		true,
		true,
		true,
		true}

	for i, s := range input {
		delim := ","
		if i >= 5 {
			delim = "/"
		}
		ret, spld := SplitString(s, delim)
		if !reflect.DeepEqual(ret, expect[i]) || spld != splited[i] {
			t.Errorf("expected :\n%q, %t\ngot\n%q, %t", expect[i], splited[i], ret, spld)
		}
	}
}
