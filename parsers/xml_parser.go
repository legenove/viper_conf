package parsers

import (
	"bytes"
	"github.com/clbanning/mxj"
	"github.com/legenove/viper"
	"github.com/spf13/afero"
	"io"
	"strings"
)

func MapInterfaceGetValue(value interface{}, isRead bool) interface{} {
	if valueMap, ok := value.(map[string]interface{}); ok {
		newMap := make(map[string]interface{})
		for k, v := range valueMap {
			if isRead && strings.HasPrefix(k, "-") {
				k = "attr_" + k[1:]
			}
			if !isRead && strings.HasPrefix(k, "attr_") {
				k = "-" + k[5:]
			}
			newMap[k] = MapInterfaceGetValue(v, isRead)
		}

		return newMap
	} else if valueSlice, ok := value.([]interface{}); ok {
		newSlice := make([]interface{}, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = MapInterfaceGetValue(v, isRead)
		}

		return newSlice
	}

	return value
}


// xml parser
type XMLParser struct {
}

func (pp *XMLParser) UnmarshalReader(v *viper.Viper, in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)
	m, err := mxj.NewMapXml(buf.Bytes(), true)
	if err != nil {
		return err
	}
	if _, ok := m["root"]; ok {
		if _m, ok := m["root"].(map[string]interface{}); ok {
			m = _m
		}
	}
	for k, v := range m {
		c[k] = MapInterfaceGetValue(v, true)
	}
	return nil
}
func (pp *XMLParser) MarshalWriter(v *viper.Viper, f afero.File, c map[string]interface{}) error {
	var m = make(map[string]interface{}, len(c))
	for k, v := range c {
		m[k] = MapInterfaceGetValue(v, false)
	}
	mv := mxj.Map(m)
	bs, err := mv.XmlIndent("", "  ", "root")
	if err != nil {
		return err
	}
	// write header
	f.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>`+"\n") )
	f.Write(bs)
	return err
}
