package develop

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"github.com/Nigel2392/versatile/internal/danger"
)

type Logger interface {
	Log(w io.Writer, a ...any) (n int64, err error)
	Logf(w io.Writer, fmt string, a ...any) (n int64, err error)
}

var (
	LOG = stackLogger{1}
	OUT = os.Stdout
)

func Log(args ...any) {
	if !DEV {
		return
	}

	LOG.Log(OUT, args...)
}

func Logf(format string, args ...any) {
	if !DEV {
		return
	}

	LOG.Logf(OUT, format, args...)
}

type stackLogger struct{ stackSize int }

const moveUp int = 3

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

		for i, name := range strBuf {
			// write func name
			buf.WriteString(name)

			// write delimiter
			if i < s.stackSize-1 {
				buf.WriteString(" / ")
			} else if writeFinalDelim {
				buf.WriteString(": ")
			}
		}
	}
}

func (s stackLogger) Log(wr io.Writer, args ...any) (n int64, err error) {
	var (
		sb  = new(bytes.Buffer)
		_sb = (*bytes.Buffer)(danger.Noescape(unsafe.Pointer(sb)))
	)

	s.writeStack(_sb, len(args) > 0)
	fmt.Fprintln(_sb, args...)

	return sb.WriteTo(wr)
}

func (s stackLogger) Logf(wr io.Writer, format string, args ...any) (n int64, err error) {
	var (
		sb  = new(bytes.Buffer)
		_sb = (*bytes.Buffer)(danger.Noescape(unsafe.Pointer(sb)))
	)

	s.writeStack(_sb, true)
	fmt.Fprintf(_sb, format, args...)
	sb.WriteRune('\n')

	return sb.WriteTo(wr)
}
