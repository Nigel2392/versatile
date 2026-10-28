package develop

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"slices"
	"strings"
	"unsafe"

	"github.com/Nigel2392/versatile/internal/danger"
)

type Logger interface {
	Log(ctx context.Context, w io.Writer, a ...any) (n int64, err error)
	Logf(ctx context.Context, w io.Writer, fmt string, a ...any) (n int64, err error)
}

var (
	LOG = stackLogger{
		stackSize:  3,
		indentBody: "   ",
		indentTail: "> ",
	}
	OUT = os.Stdout
)

type logIndentKey struct{}

func WithLogIndent(ctx context.Context, indent uint) context.Context {
	return context.WithValue(ctx, logIndentKey{}, indent)
}

func AddLogIndent(ctx context.Context, indent uint) context.Context {
	return context.WithValue(ctx, logIndentKey{}, LogIndent(ctx)+indent)
}

func LogIndent(ctx context.Context) uint {
	if ctx == nil {
		return 0
	}

	v := ctx.Value(logIndentKey{})
	if v == nil {
		return 0
	}

	return v.(uint)
}

func Log(ctx context.Context, args ...any) {
	if !DEV {
		return
	}

	LOG.Log(ctx, OUT, args...)
}

func Logf(ctx context.Context, format string, args ...any) {
	if !DEV {
		return
	}

	LOG.Logf(ctx, OUT, format, args...)
}

type stackLogger struct {
	stackSize                          int
	indentHead, indentBody, indentTail string
}

const moveUp int = 3

func (s stackLogger) writeIndent(buf *bytes.Buffer, ctx context.Context) {
	i := LogIndent(ctx)
	if i == 0 || (s.indentHead == "" && s.indentBody == "" && s.indentTail == "") {
		return
	}

	buf.Grow(len(s.indentHead) + (len(s.indentBody) * int(i)) + len(s.indentTail))

	if s.indentHead != "" {
		buf.WriteString(s.indentHead)
	}

	if s.indentBody != "" {
		for range i {
			buf.WriteString(s.indentBody)
		}
	}

	if s.indentTail != "" {
		buf.WriteString(s.indentTail)
	}
}

func (s stackLogger) writeStack(buf *bytes.Buffer, writeFinalDelim bool) {
	if s.stackSize > 0 {

		var strBuf = make([]string, 0, s.stackSize)
		for i := range s.stackSize {
			// move up the stack to get the actual calling function
			pc, _, _, ok := runtime.Caller(i + moveUp)
			details := runtime.FuncForPC(pc)

			if !ok || details == nil {
				continue
			}

			// strip pkg from func name
			funcName := details.Name()
			idx := strings.LastIndex(funcName, "/")

			// append func name
			strBuf = append(strBuf, funcName[idx+1:])
		}

		totalLen := 0
		for i, name := range strBuf {
			totalLen += len(name)
			if i < s.stackSize-1 {
				totalLen += 3
			} else if writeFinalDelim {
				totalLen += 2
			}
		}

		buf.Grow(totalLen)

		for i, name := range slices.Backward(strBuf) {
			// write func name
			buf.WriteString(name)

			// write delimiter
			if i > 0 {
				buf.WriteString(" / ")
			} else if writeFinalDelim {
				buf.WriteString(": ")
			}
		}
	}
}

func (s stackLogger) Log(ctx context.Context, wr io.Writer, args ...any) (n int64, err error) {
	var (
		sb  = new(bytes.Buffer)
		_sb = (*bytes.Buffer)(danger.Noescape(unsafe.Pointer(sb)))
	)

	s.writeIndent(_sb, ctx)
	s.writeStack(_sb, len(args) > 0)
	fmt.Fprintln(_sb, args...)

	return sb.WriteTo(wr)
}

func (s stackLogger) Logf(ctx context.Context, wr io.Writer, format string, args ...any) (n int64, err error) {
	var (
		sb  = new(bytes.Buffer)
		_sb = (*bytes.Buffer)(danger.Noescape(unsafe.Pointer(sb)))
	)

	s.writeIndent(_sb, ctx)
	s.writeStack(_sb, true)
	fmt.Fprintf(_sb, format, args...)
	sb.WriteRune('\n')

	return sb.WriteTo(wr)
}
