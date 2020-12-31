package viper_conf

import (
	"fmt"
	"github.com/legenove/viper"
	"github.com/legenove/viper_conf/parsers"
	"testing"
)

var xmlTest *ViperConf
var err error
func init() {
	viper.AddParser(&parsers.XMLParser{}, "xml")
	c, err := Tconf.Instance("test.xml", "", nil)
	if err != nil {
		xmlTest, _ = c.(*ViperConf)
	}
}

func TestXMLGetValue(t *testing.T) {
	fmt.Println(xmlTest.Conf.AllSettings())
}

