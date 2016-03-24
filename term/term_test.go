package term

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestANSIEscStrip(t *testing.T) {
	a := assert.New(t)
	var buf bytes.Buffer
	printer := NewPrinter(&Terminal{Out: &buf})
	printer.Print("1.").Styles(StyleOK).Print("Hello").Reset().Print("World")
	a.Equal("1.HelloWorld", buf.String())
}
