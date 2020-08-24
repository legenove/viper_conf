package viper_conf

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"
)

type AbcStruct struct {
	Abc string `json:"abc" mapstructure:"abc"`
}

const TestConfigPath = "./config"

var Tconf = NewConf("", TestConfigPath)

func init() {
	abc := &AbcStruct{}
	Tconf.Instance("test.json", abc, nil, nil)
}

func TestNewConf_FileChange(t *testing.T) {
	filePath := path.Join(TestConfigPath, "test.json")
	backupPath := path.Join(TestConfigPath, "test_backup.json")
	updatePath := path.Join(TestConfigPath, "test_update.json")
	v, ok := Tconf.GetViperConf("test.json")
	assert.Equal(t, true, ok)
	if !ok {
		return
	}
	s , err := v.GetString("abc")
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, "abc", s)
	// change file
	copyFile(updatePath, filePath)
	s , err = v.GetString("abc")
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, "bcd", s)
	// delete file
	os.Remove(filePath)
	s , err = v.GetString("abc")
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, "bcd", s)
	copyFile(backupPath, filePath)
	// recover first file
	s , err = v.GetString("abc")
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, "abc", s)

}

func TestNewConf(t *testing.T) {
	v, ok := Tconf.GetViperConf("test.json")
	if ok {
		fmt.Println(v.Get("abc"))
	} else {
		fmt.Println("2222")
	}
	fmt.Println(v.GetValue().(*AbcStruct))
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(1 * time.Second)
			fmt.Println("times:", i)
			fmt.Println("value", v.GetValue())
			fmt.Println("FileConf", v.GetConf())
		}
	}()
	time.Sleep(11 * time.Second)
}

func copyFile(src, dst string) {
	exec.Command("cp", src, dst).Run()
}

func scheduleDeleteAndCreateFile(filePath, backupPath string) {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			os.Remove(filePath)
			time.Sleep(1 * time.Second)
			copyFile(backupPath, filePath)
		}
	}()
}

func BenchmarkNewConf_GetValue(b *testing.B) {
	v, _ := Tconf.GetViperConf("test.json")
	filePath := path.Join(TestConfigPath, "test.json")
	backupPath := path.Join(TestConfigPath, "test_backup.json")
	scheduleDeleteAndCreateFile(filePath, backupPath)
	for i := 0; i < b.N; i++ {
		time.Sleep(10 * time.Nanosecond)
		v.GetValue()
	}
	copyFile(backupPath, filePath)
}

func BenchmarkNewConf_GetConf(b *testing.B) {
	v, _ := Tconf.GetViperConf("test.json")
	filePath := path.Join(TestConfigPath, "test.json")
	backupPath := path.Join(TestConfigPath, "test_backup.json")
	scheduleDeleteAndCreateFile(filePath, backupPath)
	for i := 0; i < b.N; i++ {
		time.Sleep(10 * time.Nanosecond)
		v.GetConf()
	}
	copyFile(backupPath, filePath)
}

func TestViperConf_GetName(t *testing.T) {
	vc := ViperConf{FileName: "a.b.c.json"}
	assert.Equal(t, "a.b.c", vc.GetName())
}

func TestViperConf_GetConfType(t *testing.T) {
	vc := ViperConf{FileName: "a.b.c.json"}
	assert.Equal(t, "json", vc.GetConfType())
}
