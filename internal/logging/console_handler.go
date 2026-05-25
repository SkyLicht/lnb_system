package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ConsoleHandler struct {
	writer io.Writer
	level  slog.Leveler
	color  bool
	attrs  []slog.Attr
	groups []string
	mutex  *sync.Mutex
}

func NewConsoleHandler(writer io.Writer, level slog.Leveler, color bool) *ConsoleHandler {
	if level == nil {
		level = slog.LevelInfo
	}

	return &ConsoleHandler{
		writer: writer,
		level:  level,
		color:  color,
		mutex:  &sync.Mutex{},
	}
}

func (h *ConsoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *ConsoleHandler) Handle(_ context.Context, record slog.Record) error {
	var builder strings.Builder

	timestamp := record.Time
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	builder.WriteString(dim(h.color, timestamp.Format("15:04:05.000")))
	builder.WriteByte(' ')
	builder.WriteString(formatLevel(record.Level, h.color))
	builder.WriteByte(' ')
	builder.WriteString(record.Message)

	for _, attr := range h.attrs {
		appendAttr(&builder, h.groups, attr, h.color)
	}
	record.Attrs(func(attr slog.Attr) bool {
		appendAttr(&builder, h.groups, attr, h.color)
		return true
	})

	builder.WriteByte('\n')

	h.mutex.Lock()
	defer h.mutex.Unlock()
	_, err := io.WriteString(h.writer, builder.String())
	return err
}

func (h *ConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := h.clone()
	next.attrs = append(next.attrs, attrs...)
	return next
}

func (h *ConsoleHandler) WithGroup(name string) slog.Handler {
	name = strings.TrimSpace(name)
	if name == "" {
		return h
	}

	next := h.clone()
	next.groups = append(next.groups, name)
	return next
}

func (h *ConsoleHandler) clone() *ConsoleHandler {
	next := *h
	next.attrs = append([]slog.Attr{}, h.attrs...)
	next.groups = append([]string{}, h.groups...)
	return &next
}

func appendAttr(builder *strings.Builder, groups []string, attr slog.Attr, color bool) {
	attr.Value = attr.Value.Resolve()
	if attr.Equal(slog.Attr{}) {
		return
	}

	keyParts := make([]string, 0, len(groups)+1)
	keyParts = append(keyParts, groups...)
	keyParts = append(keyParts, attr.Key)

	builder.WriteByte(' ')
	builder.WriteString(dim(color, strings.Join(keyParts, ".")))
	builder.WriteByte('=')
	builder.WriteString(formatValue(attr.Value))
}

func formatLevel(level slog.Level, color bool) string {
	label := strings.ToUpper(level.String())
	if len(label) < 5 {
		label += strings.Repeat(" ", 5-len(label))
	}

	if !color {
		return label
	}

	switch {
	case level >= slog.LevelError:
		return "\x1b[31;1m" + label + "\x1b[0m"
	case level >= slog.LevelWarn:
		return "\x1b[33;1m" + label + "\x1b[0m"
	case level >= slog.LevelInfo:
		return "\x1b[36;1m" + label + "\x1b[0m"
	default:
		return "\x1b[37m" + label + "\x1b[0m"
	}
}

func formatValue(value slog.Value) string {
	switch value.Kind() {
	case slog.KindString:
		return quoteIfNeeded(value.String())
	case slog.KindBool:
		return strconv.FormatBool(value.Bool())
	case slog.KindInt64:
		return strconv.FormatInt(value.Int64(), 10)
	case slog.KindUint64:
		return strconv.FormatUint(value.Uint64(), 10)
	case slog.KindFloat64:
		return strconv.FormatFloat(value.Float64(), 'f', -1, 64)
	case slog.KindDuration:
		return value.Duration().String()
	case slog.KindTime:
		return value.Time().Format(time.RFC3339)
	case slog.KindGroup:
		return formatGroup(value.Group())
	default:
		return quoteIfNeeded(fmt.Sprint(value.Any()))
	}
}

func formatGroup(attrs []slog.Attr) string {
	parts := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		attr.Value = attr.Value.Resolve()
		if attr.Equal(slog.Attr{}) {
			continue
		}
		parts = append(parts, attr.Key+"="+formatValue(attr.Value))
	}
	return "{" + strings.Join(parts, " ") + "}"
}

func quoteIfNeeded(value string) string {
	if value == "" {
		return `""`
	}
	if strings.ContainsAny(value, " \t\n\r\"") {
		return strconv.Quote(value)
	}
	return value
}

func dim(color bool, value string) string {
	if !color {
		return value
	}
	return "\x1b[2m" + value + "\x1b[0m"
}
