package term

const (
	StyleOK   = "ok"
	StyleErr  = "err"
	StyleWarn = "warn"
	StyleAsk  = "ask"
	StyleHi   = "hi"
	StyleLo   = "lo"
	StyleEm   = "em"
	StyleB    = "b"
	StyleI    = "i"
)

var (
	DefaultPalette = Palette{
		Stylers: map[string]Styler{
			"bold":         Colorizer("1"),
			"dim":          Colorizer("2"),
			"underline":    Colorizer("4"),
			"blinkslow":    Colorizer("5"),
			"blinkfast":    Colorizer("6"),
			"invert":       Colorizer("7"),
			"hidden":       Colorizer("8"),
			"black":        Colorizer("30"),
			"red":          Colorizer("31"),
			"green":        Colorizer("32"),
			"yellow":       Colorizer("33"),
			"blue":         Colorizer("34"),
			"magenta":      Colorizer("35"),
			"cyan":         Colorizer("36"),
			"lightgray":    Colorizer("37"),
			"darkgray":     Colorizer("90"),
			"lightred":     Colorizer("91"),
			"lightgreen":   Colorizer("92"),
			"lightyellow":  Colorizer("93"),
			"lightblue":    Colorizer("94"),
			"lightmagenta": Colorizer("95"),
			"lightcyan":    Colorizer("96"),
			"white":        Colorizer("97"),
			"reset":        Colorizer("0"),
			"resetbold":    Colorizer("21"),

			"ok":   AliasStyler("green"),
			"err":  AliasStyler("red"),
			"warn": AliasStyler("yellow"),
			"hi":   AliasStyler("white"),
			"lo":   AliasStyler("darkgray"),
			"em":   AliasStyler("invert"),
			"b":    AliasStyler("bold"),
			"i":    AliasStyler("underline"),
			"ask":  AliasStyler("hi"),
		},
	}
)

type Styler func(*Palette, string) string

// Palette defines a set of named stylers
type Palette struct {
	Stylers map[string]Styler
}

func PrefixStyler(prefix string) Styler {
	return func(pal *Palette, str string) string {
		return prefix + str
	}
}

func SuffixStyler(suffix string) Styler {
	return func(pal *Palette, str string) string {
		return str + suffix
	}
}

func ANSIStyler(seqs ...string) Styler {
	prefix := ""
	for _, seq := range seqs {
		prefix += "\x1b\x5b" + seq
	}
	return PrefixStyler(prefix)
}

func ResetStyler() Styler {
	return SuffixStyler("\x1b\x5b0m")
}

func Colorizer(codes ...string) Styler {
	seqs := make([]string, 0, len(codes))
	for _, code := range codes {
		seqs = append(seqs, code+"m")
	}
	return ANSIStyler(seqs...)
}

func AliasStyler(name string) Styler {
	return func(pal *Palette, str string) string {
		return pal.Apply(str, name)
	}
}

func StackStyler(stylers ...Styler) Styler {
	return func(pal *Palette, str string) string {
		for _, s := range stylers {
			str = s(pal, str)
		}
		return str
	}
}

func (p *Palette) Apply(msg string, names ...string) string {
	for _, name := range names {
		if styler, exist := p.Stylers[name]; exist && styler != nil {
			msg = styler(p, msg)
		}
	}
	return msg
}
