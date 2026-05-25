package envfile

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Load(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("%s:%d: expected KEY=value", path, lineNumber)
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("%s:%d: key is required", path, lineNumber)
		}

		values[key] = trimValue(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return values, nil
}

func Required(values map[string]string, key string) (string, error) {
	value := strings.TrimSpace(values[key])
	if value == "" {
		return "", fmt.Errorf("%s is required in .env", key)
	}
	return value, nil
}

func trimValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) < 2 {
		return value
	}

	if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
		(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
		return value[1 : len(value)-1]
	}

	return value
}
