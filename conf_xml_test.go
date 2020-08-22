package viper_conf

import (
	"bytes"
	"fmt"
	"github.com/legenove/viper"
	"github.com/legenove/viper_conf/parsers"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
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

