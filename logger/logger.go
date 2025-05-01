package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"sync"
)

const (
	LevelAi slog.Level = slog.LevelInfo + 1
)

const (
	ansiDim   = "\x1b[2m" // Dim text
	ansiReset = "\x1b[0m" // Reset all formatting
)

const (
	aiLogStartMarker = "--- AI LOG START ---"
	aiLogEndMarker   = "--- AI LOG END ---"
)

var defaultLogger *slog.Logger
var once sync.Once

func Init(level slog.Level) {
	once.Do(func() {
		opts := &slog.HandlerOptions{
			Level: level, // Set the minimum level for logs to be processed
		}
		handler := NewAiHandler(os.Stdout, opts)
		defaultLogger = slog.New(handler)
		slog.SetDefault(defaultLogger)
	})
}

func Get() *slog.Logger {
	if defaultLogger == nil {
		panic("logger: Init must be called before Get")
	}
	return defaultLogger
}

type AiHandler struct {
	next            slog.Handler                        // The underlying handler (e.g., TextHandler)
	mu              *sync.Mutex                         // Mutex to protect writes to the shared output buffer
	out             io.Writer                           // The original output writer (e.g., os.Stdout)
	replaceAttrFunc func([]string, slog.Attr) slog.Attr // Store the ReplaceAttr function
}

func NewAiHandler(out io.Writer, opts *slog.HandlerOptions) *AiHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	if opts.Level == nil {
		opts.Level = slog.LevelInfo
	}

	effectiveReplaceAttr := replaceLevelNames(opts.ReplaceAttr)

	handlerOpts := *opts
	handlerOpts.ReplaceAttr = effectiveReplaceAttr

	return &AiHandler{
		next:            slog.NewTextHandler(out, &handlerOpts),
		mu:              &sync.Mutex{},        // Mutex for safe concurrent access to the internal buffer used in Handle
		out:             out,                  // Store the original output writer
		replaceAttrFunc: effectiveReplaceAttr, // Store the function for later use in Handle
	}
}

func (h *AiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *AiHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level == LevelAi {
		h.mu.Lock() // Protect access when formatting and writing
		defer h.mu.Unlock()

		buf := new(bytes.Buffer)

		tempHandlerOpts := &slog.HandlerOptions{
			Level:       slog.LevelDebug,   // Ensure the temp handler processes AI level
			ReplaceAttr: h.replaceAttrFunc, // Use the stored ReplaceAttr function
		}
		tempHandler := slog.NewTextHandler(buf, tempHandlerOpts)

		recordClone := r.Clone()
		if err := tempHandler.Handle(ctx, recordClone); err != nil {
			return fmt.Errorf("failed to handle AI log record in temp buffer: %w", err)
		}

		_, err := fmt.Fprintf(h.out, "%s%s\n%s%s%s\n", ansiDim, aiLogStartMarker, buf.String(), aiLogEndMarker, ansiReset)
		if err != nil {
			return fmt.Errorf("failed to write formatted AI log: %w", err)
		}
		return nil
	}

	return h.next.Handle(ctx, r)
}

func (h *AiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &AiHandler{next: h.next.WithAttrs(attrs), mu: h.mu, out: h.out, replaceAttrFunc: h.replaceAttrFunc}
}

func (h *AiHandler) WithGroup(name string) slog.Handler {
	return &AiHandler{next: h.next.WithGroup(name), mu: h.mu, out: h.out, replaceAttrFunc: h.replaceAttrFunc}
}

func replaceLevelNames(next func([]string, slog.Attr) slog.Attr) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.LevelKey {
			level, ok := a.Value.Any().(slog.Level)
			if ok {
				switch level {
				case LevelAi:
					a.Value = slog.StringValue("AI")
				case slog.LevelDebug:
					a.Value = slog.StringValue("DEBUG")
				case slog.LevelInfo:
					a.Value = slog.StringValue("INFO")
				case slog.LevelWarn:
					a.Value = slog.StringValue("WARN")
				case slog.LevelError:
					a.Value = slog.StringValue("ERROR")
				default:
					a.Value = slog.StringValue(level.String())
				}
			}
		}
		if next != nil {
			a = next(groups, a)
		}
		return a
	}
}

func Debug(msg string, args ...any) {
	Get().Log(context.Background(), slog.LevelDebug, msg, args...)
}

func Info(msg string, args ...any) {
	Get().Log(context.Background(), slog.LevelInfo, msg, args...)
}

func Warn(msg string, args ...any) {
	Get().Log(context.Background(), slog.LevelWarn, msg, args...)
}

func Error(msg string, args ...any) {
	Get().Log(context.Background(), slog.LevelError, msg, args...)
}

func Ai(msg string, args ...any) {
	Get().Log(context.Background(), LevelAi, msg, args...)
}

func Log(level slog.Level, msg string, args ...any) {
	Get().Log(context.Background(), level, msg, args...)
}

func Aif(fn func()) {
	envValue := os.Getenv("LEMC_AI_FUNC")
	if enabled, _ := strconv.ParseBool(envValue); enabled {
		fn()
	}
}
