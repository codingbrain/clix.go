package gen

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

const (
	// DefaultIndentSize is the default value for new indent size
	DefaultIndentSize = 4
	// DefaultIndentChar is the default char for indent
	DefaultIndentChar = " "
)

// Writer is the output for code generation with indent support
type Writer struct {
	Output     io.Writer
	IndentSize int
	IndentChar string

	level  int
	indent string
}

type nonCloserWriter struct {
	writer io.Writer
}

func (w *nonCloserWriter) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}

// Writeln formats the output with indent prefixed
func (w *Writer) Writeln(format string, args ...interface{}) {
	fmt.Fprintf(w.Output, w.indentSpaces()+format+"\n", args...)
}

// Indent creates a nested writer with extended indent
func (w *Writer) Indent() *Writer {
	return &Writer{
		Output:     &nonCloserWriter{writer: w.Output},
		IndentSize: w.IndentSize,
		level:      w.level + 1,
	}
}

// Close closes the output if it's io.Closer
func (w *Writer) Close() error {
	if closer, ok := w.Output.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (w *Writer) indentSpaces() string {
	if w.indent == "" && w.IndentSize > 0 && w.level > 0 {
		indent := ""
		chars := w.IndentChar
		if chars == "" {
			chars = DefaultIndentChar
		}
		for i := 0; i < w.IndentSize; i++ {
			indent += chars
		}
		for i := 0; i < w.level; i++ {
			w.indent += indent
		}
	}
	return w.indent
}

// NewFileWriter creates a file backed Writer
func NewFileWriter(filename string) (w *Writer, err error) {
	w = &Writer{IndentSize: DefaultIndentSize}
	if filename == "" || filename == "-" {
		w.Output = &nonCloserWriter{writer: os.Stdout}
	} else {
		os.MkdirAll(filepath.Dir(filename), 0755)
		w.Output, err = os.OpenFile(filename,
			syscall.O_WRONLY|syscall.O_CREAT|syscall.O_TRUNC, 0644)
	}
	return
}
