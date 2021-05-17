package client

import (
	"bufio"
	"os"
	"strings"
	"unicode"
)

type propertyReader map[string]string

func newPropertyReader(filePath string) (propertyReader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	props := make(propertyReader)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		kvp := strings.SplitN(scanner.Text(), "=", 2)
		key, value := strings.TrimSpace(kvp[0]), ""
		if len(kvp) > 1 {
			value = strings.TrimLeftFunc(kvp[1], unicode.IsSpace)
		}
		if len(key) > 0 && key[0] == '#' {
			continue
		}
		props[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return props, nil
}

func (pr propertyReader) getString(key string, value *string) bool {
	if str, ok := pr[key]; ok {
		*value = str
		return true
	}
	return false
}
