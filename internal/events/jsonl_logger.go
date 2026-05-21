package events

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type JSONLLogger struct {
	encoder *json.Encoder
	file    *os.File
	mutex   sync.Mutex
}

func NewJSONLLogger(logFile string, stdout io.Writer) (*JSONLLogger, error) {
	var writer io.Writer = stdout
	var file *os.File

	if strings.TrimSpace(logFile) != "" {
		directory := filepath.Dir(logFile)
		if directory != "." {
			if err := os.MkdirAll(directory, 0755); err != nil {
				return nil, err
			}
		}

		openedFile, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}

		file = openedFile
		writer = io.MultiWriter(stdout, file)
	}

	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)

	return &JSONLLogger{
		encoder: encoder,
		file:    file,
	}, nil
}

func (l *JSONLLogger) Write(event ChangeEvent) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.encoder.Encode(event)
}

func (l *JSONLLogger) Close() error {
	if l.file == nil {
		return nil
	}

	return l.file.Close()
}
