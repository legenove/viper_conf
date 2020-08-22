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
	xmlTest, err = Tconf.Instance("test.xml", nil, nil, nil)
}

func TestXMLGetValue(t *testing.T) {
	fmt.Println(xmlTest.Conf.AllSettings())
}

