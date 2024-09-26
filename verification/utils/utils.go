package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func PrettyPrintJson(v interface{}) string {
	j, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	str := string(j)
	fmt.Print(str)
	return str
}

func SplitString(str string, delim string) ([]string, bool) {
	var ret []string
	for _, s := range strings.Split(str, delim) {
		ret = append(ret, strings.Trim(s, " "))
	}
	return ret, len(ret) > 1
}

func ReadFile(path string) *os.File {
	reader, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	return reader
}
