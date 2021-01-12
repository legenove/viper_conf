package viper_conf

import (
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/legenove/easyconfig/ifacer"
	"github.com/legenove/viper"
	"github.com/legenove/viper_conf/parsers"
)

func init() {
	viper.AddParser(&parsers.XMLParser{}, "xml")
}

type FileConf struct {
	Env         string
	Path        string
	DefaultPath string
	Val         map[string]*ViperConf
	sync.RWMutex
}

func NewConf(env string, defaultPath string) *FileConf {
	path := defaultPath
	if env != "" {
		path = filepath.Join(defaultPath, env)
	}
	createDir(path)
	return &FileConf{
		Env:         env,
		Path:        path,
		DefaultPath: defaultPath,
		Val:         make(map[string]*ViperConf),
	}
}

func (c *FileConf) RegisterViperConf(viperConf *ViperConf) {
	viperConf.setBaseConf(c)
	c.Lock()
	defer c.Unlock()
	c.Val[viperConf.GetName()] = viperConf
}

func (c *FileConf) GetViperConf(filename string) (*ViperConf, bool) {
	c.RLock()
	defer c.RUnlock()
	v, ok := c.Val[filename]
	return v, ok
}

func (c *FileConf) Instance(fileName, _type string, val interface{}, opts ...ifacer.OptionFunc) (ifacer.Configer, error) {
	c.Lock()
	defer c.Unlock()
	var v *ViperConf
	var ok bool
	v, ok = c.Val[fileName]
	if ok {
		return v, nil
	}
	v = NewViperConfig(fileName)
	v.SetConfType(_type)
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(v)
	}
	x, err := parseConf(c, v.GetName(), v.GetConfType())
	if err != nil {
		return nil, err
	}
	v.Val = val
	v.HasViper = true
	v.Conf = x
	v.setBaseConf(c)
	c.Val[fileName] = v
	return v, nil
}

// 必须需要注册到基础配置才能进行读取配置
type ViperConf struct {
	BaseConf     *FileConf         // 全局配置
	ConfLock     sync.Mutex        // 锁
	ValLock      sync.Mutex        // 锁
	LoadLock     sync.Mutex        // 锁
	FileName     string            // 文件名
	Type         string            // 文件类型
	HasVal       bool              // 是否有值
	Val          interface{}       // 配置的值
	OnChange     chan struct{}     // 变更通知
	onChangeFunc ifacer.ChangeFunc // 改变 func
	onRemoveFunc ifacer.ChangeFunc // 删除 func
	readingViper bool              // 是否正在读配置
	HasViper     bool              // 是否有配置
	Conf         *viper.Viper      // 配置对象
	Error        error
}

func NewViperConfig(fileName string) *ViperConf {
	return &ViperConf{
		OnChange: make(chan struct{}, 8),
		FileName: fileName,
	}
}

func (vc *ViperConf) OnChangeFunc(iv ifacer.Configer) {
	if vc.onChangeFunc != nil {
		vc.onChangeFunc(iv)
	}
	DefaultOnChangeFunc(iv)
}

func (vc *ViperConf) OnRemoveFunc(iv ifacer.Configer) {
	if vc.onRemoveFunc != nil {
		vc.onRemoveFunc(iv)
	}
	DefaultOnRemoveFunc(iv)
}

func (vc *ViperConf) GetName() string {
	names := strings.Split(vc.FileName, ".")
	return strings.Join(names[:len(names)-1], ".")
}

func (vc *ViperConf) GetFullName() string {
	return vc.FileName
}

func (vc *ViperConf) GetConfType() string {
	if len(vc.Type) == 0 {
		names := strings.Split(vc.FileName, ".")
		vc.Type = names[len(names)-1]
	}
	return vc.Type
}

func (vc *ViperConf) SetConfType(t string) {
	vc.Type = t
}

func (vc *ViperConf) setBaseConf(conf2 *FileConf) {
	vc.BaseConf = conf2
}

func (vc *ViperConf) SetValue(val interface{}) *ViperConf {
	vc.Val = val
	return vc
}

func (vc *ViperConf) SetOnChangeFunc(onChangeFunc ifacer.ChangeFunc) {
	vc.onChangeFunc = onChangeFunc
}

func (vc *ViperConf) SetOnRemoveFunc(onRemoveFunc ifacer.ChangeFunc) {
	vc.onRemoveFunc = onRemoveFunc
}

func (vc *ViperConf) preseConf() {
	if !vc.HasViper {
		vc.ConfLock.Lock()
		defer vc.ConfLock.Unlock()
		if vc.HasViper {
			return
		}
		x, err := parseConf(vc.BaseConf, vc.GetName(), vc.GetConfType())
		if err != nil {
			vc.Error = err
			return
		}
		vc.Error = nil
		vc.HasViper = true
		vc.Conf = x
	}
}

func (vc *ViperConf) UnmarshalKey(key string, rawVal interface{}) error {
	conf := vc.GetConf()
	if conf == nil {
		return vc.Error
	}
	return conf.UnmarshalKey(key, rawVal)
}

func (vc *ViperConf) Unmarshal(rawVal interface{}) error {
	conf := vc.GetConf()
	if conf == nil {
		return vc.Error
	}
	return conf.Unmarshal(rawVal)
}

func (vc *ViperConf) GetConf() *viper.Viper {
	if !vc.HasViper {
		vc.preseConf()
		// 从新加载配置
		if vc.HasViper {
			vc.LoadLock.Lock()
			if vc.HasViper {
				vc.OnChangeFunc(vc)
			}
			vc.LoadLock.Unlock()
		}
	}
	return vc.Conf
}

func (vc *ViperConf) GetValue() interface{} {
	if !vc.HasViper {
		// 没有配置时，要加载配置
		vc.GetConf()
	}
	if vc.Val != nil && !vc.HasVal {
		vc.LoadLock.Lock()
		if vc.Val != nil && !vc.HasVal {
			vc.OnChangeFunc(vc)
		}
		vc.LoadLock.Unlock()
	}
	return vc.Val
}

func (vc *ViperConf) Get(key string) (interface{}, error) {
	conf := vc.GetConf()
	if conf == nil {
		return "", vc.Error
	}
	return conf.Get(key), nil
}

func (vc *ViperConf) GetString(key string) (string, error) {
	conf := vc.GetConf()
	if conf == nil {
		return "", vc.Error
	}
	return conf.GetString(key), nil
}

// GetBool returns the value associated with the key as a boolean.
func (vc *ViperConf) GetBool(key string) (bool, error) {
	conf := vc.GetConf()
	if conf == nil {
		return false, vc.Error
	}
	return conf.GetBool(key), nil
}

// GetInt returns the value associated with the key as an integer.
func (vc *ViperConf) GetInt(key string) (int, error) {
	conf := vc.GetConf()
	if conf == nil {
		return 0, vc.Error
	}
	return conf.GetInt(key), nil
}

// GetInt32 returns the value associated with the key as an integer.
func (vc *ViperConf) GetInt32(key string) (int32, error) {
	conf := vc.GetConf()
	if conf == nil {
		return 0, vc.Error
	}
	return conf.GetInt32(key), nil
}

// GetInt64 returns the value associated with the key as an integer.
func (vc *ViperConf) GetInt64(key string) (int64, error) {
	conf := vc.GetConf()
	if conf == nil {
		return 0, vc.Error
	}
	return conf.GetInt64(key), nil
}

// GetFloat64 returns the value associated with the key as a float64.
func (vc *ViperConf) GetFloat64(key string) (float64, error) {
	conf := vc.GetConf()
	if conf == nil {
		return 0, vc.Error
	}
	return conf.GetFloat64(key), nil
}

// GetTime returns the value associated with the key as time.
func (vc *ViperConf) GetTime(key string) (time.Time, error) {
	conf := vc.GetConf()
	if conf == nil {
		return time.Time{}, vc.Error
	}
	return conf.GetTime(key), nil
}

// GetDuration returns the value associated with the key as a duration.
func (vc *ViperConf) GetDuration(key string) (time.Duration, error) {
	conf := vc.GetConf()
	if conf == nil {
		return 0, vc.Error
	}
	return conf.GetDuration(key), nil
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (vc *ViperConf) GetStringSlice(key string) ([]string, error) {
	conf := vc.GetConf()
	if conf == nil {
		return nil, vc.Error
	}
	return conf.GetStringSlice(key), nil
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func (vc *ViperConf) GetStringMap(key string) (map[string]interface{}, error) {
	conf := vc.GetConf()
	if conf == nil {
		return nil, vc.Error
	}
	return conf.GetStringMap(key), nil
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (vc *ViperConf) GetStringMapString(key string) (map[string]string, error) {
	conf := vc.GetConf()
	if conf == nil {
		return nil, vc.Error
	}
	return conf.GetStringMapString(key), nil
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func (vc *ViperConf) GetStringMapStringSlice(key string) (map[string][]string, error) {
	conf := vc.GetConf()
	if conf == nil {
		return nil, vc.Error
	}
	return conf.GetStringMapStringSlice(key), nil
}

// GetSizeInBytes returns the size of the value associated with the given key
// in bytes.
func (vc *ViperConf) GetSizeInBytes(key string) (uint, error) {
	conf := vc.GetConf()
	if conf == nil {
		return 0, vc.Error
	}
	return conf.GetSizeInBytes(key), nil
}

// GetSizeInBytes returns the size of the value associated with the given key
// in bytes.
func (vc *ViperConf) AllKeys() []string {
	conf := vc.GetConf()
	if conf == nil {
		return []string{}
	}
	return conf.AllKeys()
}

func (vc *ViperConf) OnChangeChan() <-chan struct{} {
	return vc.OnChange
}

func parseConf(c *FileConf, name, confType string) (*viper.Viper, error) {
	x := viper.New()
	x.SetConfigName(name)
	x.SetConfigType(confType)
	x.AddConfigPath(c.Path)
	if c.DefaultPath != "" {
		x.AddConfigPath(c.DefaultPath)
	}
	if err := x.ReadInConfig(); err != nil {
		return nil, err
	}
	x.WatchConfig()
	x.OnConfigChange(func(e fsnotify.Event) {
		confChangePool(getFileName(name, confType), c)
	})
	x.OnConfigRemove(func(e fsnotify.Event) {
		confRemovePool(getFileName(name, confType), c)
	})
	return x, nil
}

func confChangePool(key string, c *FileConf) {
	v, ok := c.Val[key]
	if !ok {
		return
	}
	v.OnChangeFunc(v)
}

func confRemovePool(key string, c *FileConf) {
	v, ok := c.Val[key]
	if !ok {
		return
	}
	v.OnRemoveFunc(v)
}

func DefaultOnChangeFunc(iv ifacer.Configer) {
	v, ok := iv.(*ViperConf)
	if !ok {
		return
	}
	if v.Val != nil && v.Conf != nil {
		v.ValLock.Lock()
		defer v.ValLock.Unlock()
		if v.Conf == nil {
			return
		}
		err := v.Conf.Unmarshal(&v.Val)
		if err != nil {
			return
		}
		v.HasVal = true
	}
	go func() {
		// 防止无人使用 onchange channel
		if len(v.OnChange) == 0 {
			select {
			case v.OnChange <- struct{}{}:
			case <-time.After(time.Millisecond):
			}
		}
	}()
}

func DefaultOnRemoveFunc(iv ifacer.Configer) {
	v, ok := iv.(*ViperConf)
	if !ok {
		return
	}
	if v.Conf != nil {
		v.ConfLock.Lock()
		defer v.ConfLock.Unlock()
		if v.Conf == nil {
			return
		}
		v.HasViper = false
	}
}

func getFileName(name, confType string) string {
	return concatenateStrings(name, ".", confType)
}
