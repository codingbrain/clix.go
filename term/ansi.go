package term

import (
	"bytes"
)

const (
	EscEndMin = byte(64)
	EscEndMax = byte(126)
)

var (
	EscPrefix = []byte{'\x1b', '\x5b'}
)

type ANSIEscStrip struct {
	buf bytes.Buffer
	seq int
}

func (s *ANSIEscStrip) Write(p []byte) (int, error) {
	l := len(p)
	for i := 0; i < l; i++ {
		var w []byte
		if s.seq == 0 {
			pos := bytes.IndexByte(p[i:], EscPrefix[0])
			if pos >= 0 {
				w = p[i : i+pos]
				s.seq++
				i += pos
			} else {
				w = p[i:]
				i = l
			}
		} else if s.seq < len(EscPrefix) {
			if p[i] == EscPrefix[s.seq] {
				s.seq++
			} else {
				w = EscPrefix[0:s.seq]
				s.seq = 0
				i--
			}
		} else if p[i] >= EscEndMin && p[i] <= EscEndMax {
			s.seq = 0
		}
		if len(w) > 0 {
			if _, err := s.buf.Write(w); err != nil {
				return i, err
			}
		}
	}
	return l, nil
}

func (s *ANSIEscStrip) String() string {
	return s.buf.String()
}

func (s *ANSIEscStrip) Bytes() []byte {
	return s.buf.Bytes()
}

func (s *ANSIEscStrip) Reset() {
	s.buf = bytes.Buffer{}
	s.seq = 0
}
