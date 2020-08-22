package parsers

import (
	"bytes"
	"github.com/legenove/viper"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func init() {
	viper.AddParser(&XMLParser{}, "xml")
}

var xmlExample = []byte(`<?xml version="1.0" encoding="utf-8"?>
<root>
    <a>a</a>
    <b>b</b>
    <c>c</c>
    <d dd="dd" dd2="dd2"/>
</root>`)

var xmlWriteExpected = []byte(`<?xml version="1.0" encoding="utf-8"?>
<root>
  <a>a</a>
  <b>b</b>
  <c>c</c>
  <d dd="dd" dd2="dd2"/>
</root>`)

func TestXML(t *testing.T) {
	v := viper.New()
	v.SetConfigType("xml")
	r := strings.NewReader(string(xmlExample))
	err := v.ReadConfig(r)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "dd", v.Get("d.attr_dd"))
}

func TestWriteConfig(t *testing.T) {
	fs := afero.NewMemMapFs()
	testCases := map[string]struct {
		configName      string
		inConfigType    string
		outConfigType   string
		fileName        string
		input           []byte
		expectedContent []byte
	}{
		"xml with file extension": {
			configName:      "c",
			inConfigType:    "xml",
			outConfigType:   "xml",
			fileName:        "c.xml",
			input:           xmlExample,
			expectedContent: xmlWriteExpected,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.SetFs(fs)
			v.SetConfigName(tc.fileName)
			v.SetConfigType(tc.inConfigType)

			err := v.ReadConfig(bytes.NewBuffer(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			v.SetConfigType(tc.outConfigType)
			if err := v.WriteConfigAs(tc.fileName); err != nil {
				t.Fatal(err)
			}
			read, err := afero.ReadFile(fs, tc.fileName)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.expectedContent, read)
		})
	}
}
