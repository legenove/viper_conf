package viper_conf

import (
	"fmt"
	"github.com/legenove/viper"
	"github.com/legenove/viper_conf/parsers"
	"testing"
)

var xmlTest *ViperConf
var err error
var ok bool

func init() {
	fmt.Println(1111)
	viper.AddParser(&parsers.XMLParser{}, "xml")
	c, err := Tconf.Instance("test.xml", "xml", nil)
	if err == nil {
		xmlTest, ok = c.(*ViperConf)
	}
}

func TestXMLGetValue(t *testing.T) {
	fmt.Println(xmlTest.Conf)
	fmt.Println(xmlTest.Conf.AllSettings())
}
