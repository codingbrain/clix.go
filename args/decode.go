package args

import (
	"io"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func DecodeCmds(reader io.Reader) (*Command, error) {
	if def, err := ioutil.ReadAll(reader); err != nil {
		return nil, err
	} else {
		return DecodeCmdsBytes(def)
	}
}

func DecodeCmdsBytes(def []byte) (*Command, error) {
	cmd := &Command{}
	if err := yaml.Unmarshal(def, cmd); err != nil {
		return nil, err
	} else {
		return cmd, cmd.Normalize()
	}
}

func DecodeCmdsString(def string) (*Command, error) {
	return DecodeCmdsBytes([]byte(def))
}

func DecodeCliDef(reader io.Reader) (*CliDef, error) {
	if def, err := ioutil.ReadAll(reader); err != nil {
		return nil, err
	} else {
		return DecodeCliDefBytes(def)
	}
}

func DecodeCliDefFile(filename string) (*CliDef, error) {
	if f, err := os.Open(filename); err != nil {
		return nil, err
	} else {
		defer f.Close()
		return DecodeCliDef(f)
	}
}

func DecodeCliDefBytes(def []byte) (*CliDef, error) {
	cliDef := &CliDef{}
	if err := yaml.Unmarshal(def, cliDef); err != nil {
		return nil, err
	} else {
		return cliDef, cliDef.Normalize()
	}
}

func DecodeCliDefString(def string) (*CliDef, error) {
	return DecodeCliDefBytes([]byte(def))
}
