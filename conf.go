package viper_conf

import (
	"github.com/fsnotify/fsnotify"
	"github.com/legenove/utils"
	"github.com/legenove/viper"
	"github.com/legenove/viper_conf/parsers"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func init() {
	viper.AddParser(&parsers.XMLParser{}, "xml")
}

type conf struct {
	Env         string
	Path        string
	DefaultPath string
	Val         map[string]*ViperConf
	sync.RWMutex
}

func NewConf(env string, defaultPath string) *conf {
	path := defaultPath
	if env != "" {
		path = filepath.Join(defaultPath, env)
	}
	utils.CreateDir(path)
	return &conf{
		Env:         env,
		Path:        path,
		DefaultPath: defaultPath,
		Val:         make(map[string]*ViperConf),
	}
}

func (c *conf) RegisterViperConf(viperConf *ViperConf) {
	viperConf.setBaseConf(c)
	c.Lock()
	defer c.Unlock()
	c.Val[viperConf.GetName()] = viperConf
}

func (c *conf) GetViperConf(filename string) (*ViperConf, bool) {
	c.RLock()
	defer c.RUnlock()
	v, ok := c.Val[filename]
	return v, ok
}

func (c *conf) Instance(fileName string, val interface{},
	onChangeFunc func(*ViperConf), onRemoveFunc func(*ViperConf)) (*ViperConf, error) {
	c.Lock()
	defer c.Unlock()
	var v *ViperConf
	var ok bool
	v, ok = c.Val[fileName]
	if ok {
		return v, nil
	}
	v = NewViperConfig(fileName)
	if onChangeFunc != nil {
		v.SetOnChangeFunc(onChangeFunc)
	}
	if onRemoveFunc != nil {
		v.SetOnRemoveFunc(onRemoveFunc)
	}
	x, err := parseConf(c, v.GetName(), v.GetConfType(), onChangeFunc, onRemoveFunc)
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
	BaseConf     *conf            // 全局配置
	ConfLock     sync.Mutex       // 锁
	ValLock      sync.Mutex       // 锁
	LoadLock     sync.Mutex       // 锁
	FileName     string           // 文件名
	HasVal       bool             // 是否有值
	Val          interface{}      // 配置的值
	OnChangeFunc func(*ViperConf) // 改变 func
	OnRemoveFunc func(*ViperConf) // 删除 func
	readingViper bool             // 是否正在读配置
	HasViper     bool             // 是否有配置
	Conf         *viper.Viper     // 配置对象
	Error        error
}

func NewViperConfig(fileName string) *ViperConf {
	return &ViperConf{
		FileName:     fileName,
		OnChangeFunc: DefaultOnChangeFunc,
		OnRemoveFunc: DefaultOnRemoveFunc,
	}
}

func (vc *ViperConf) GetName() string {
	names := strings.Split(vc.FileName, ".")
	return strings.Join(names[:len(names)-1], ".")
}

func (vc *ViperConf) GetConfType() string {
	names := strings.Split(vc.FileName, ".")
	return names[len(names)-1]
}

func (vc *ViperConf) setBaseConf(conf2 *conf) {
	vc.BaseConf = conf2
}

func (vc *ViperConf) SetValue(val interface{}) *ViperConf {
	vc.Val = val
	return vc
}

func (vc *ViperConf) SetOnChangeFunc(onChangeFunc func(*ViperConf)) {
	vc.OnChangeFunc = onChangeFunc
}

func (vc *ViperConf) SetOnRemoveFunc(onRemoveFunc func(*ViperConf)) {
	vc.OnRemoveFunc = onRemoveFunc
}

func (vc *ViperConf) preseConf() {
	if !vc.HasViper {
		vc.ConfLock.Lock()
		defer vc.ConfLock.Unlock()
		if vc.HasViper {
			return
		}
		x, err := parseConf(vc.BaseConf, vc.GetName(), vc.GetConfType(), vc.OnChangeFunc, vc.OnRemoveFunc)
		if err != nil {
			vc.Error = err
			return
		}
		vc.Error = nil
		vc.HasViper = true
		vc.Conf = x
	}
}

func (vc *ViperConf) GetConf() *viper.Viper {
	if !vc.HasViper {
		vc.preseConf()
		// 从新加载配置
		if vc.HasViper && vc.OnChangeFunc != nil {
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
	if vc.Val != nil && !vc.HasVal && vc.OnChangeFunc != nil {
		vc.LoadLock.Lock()
		if vc.Val != nil && !vc.HasVal {
			vc.OnChangeFunc(vc)
		}
		vc.LoadLock.Unlock()
	}
	return vc.Val
}

func (cb *ViperConf) Get(key string) (interface{}, error) {
	conf := cb.GetConf()
	if conf == nil {
		return "", cb.Error
	}
	return conf.Get(key), nil
}

func (cb *ViperConf) GetString(key string) (string, error) {
	conf := cb.GetConf()
	if conf == nil {
		return "", cb.Error
	}
	return conf.GetString(key), nil
}

// GetBool returns the value associated with the key as a boolean.
func (cb *ViperConf) GetBool(key string) (bool, error) {
	conf := cb.GetConf()
	if conf == nil {
		return false, cb.Error
	}
	return conf.GetBool(key), nil
}

// GetInt returns the value associated with the key as an integer.
func (cb *ViperConf) GetInt(key string) (int, error) {
	conf := cb.GetConf()
	if conf == nil {
		return 0, cb.Error
	}
	return conf.GetInt(key), nil
}

// GetInt32 returns the value associated with the key as an integer.
func (cb *ViperConf) GetInt32(key string) (int32, error) {
	conf := cb.GetConf()
	if conf == nil {
		return 0, cb.Error
	}
	return conf.GetInt32(key), nil
}

// GetInt64 returns the value associated with the key as an integer.
func (cb *ViperConf) GetInt64(key string) (int64, error) {
	conf := cb.GetConf()
	if conf == nil {
		return 0, cb.Error
	}
	return conf.GetInt64(key), nil
}

// GetFloat64 returns the value associated with the key as a float64.
func (cb *ViperConf) GetFloat64(key string) (float64, error) {
	conf := cb.GetConf()
	if conf == nil {
		return 0, cb.Error
	}
	return conf.GetFloat64(key), nil
}

// GetTime returns the value associated with the key as time.
func (cb *ViperConf) GetTime(key string) (time.Time, error) {
	conf := cb.GetConf()
	if conf == nil {
		return time.Time{}, cb.Error
	}
	return conf.GetTime(key), nil
}

// GetDuration returns the value associated with the key as a duration.
func (cb *ViperConf) GetDuration(key string) (time.Duration, error) {
	conf := cb.GetConf()
	if conf == nil {
		return 0, cb.Error
	}
	return conf.GetDuration(key), nil
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (cb *ViperConf) GetStringSlice(key string) ([]string, error) {
	conf := cb.GetConf()
	if conf == nil {
		return nil, cb.Error
	}
	return conf.GetStringSlice(key), nil
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func (cb *ViperConf) GetStringMap(key string) (map[string]interface{}, error) {
	conf := cb.GetConf()
	if conf == nil {
		return nil, cb.Error
	}
	return conf.GetStringMap(key), nil
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (cb *ViperConf) GetStringMapString(key string) (map[string]string, error) {
	conf := cb.GetConf()
	if conf == nil {
		return nil, cb.Error
	}
	return conf.GetStringMapString(key), nil
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func (cb *ViperConf) GetStringMapStringSlice(key string) (map[string][]string, error) {
	conf := cb.GetConf()
	if conf == nil {
		return nil, cb.Error
	}
	return conf.GetStringMapStringSlice(key), nil
}

// GetSizeInBytes returns the size of the value associated with the given key
// in bytes.
func (cb *ViperConf) GetSizeInBytes(key string) (uint, error) {
	conf := cb.GetConf()
	if conf == nil {
		return 0, cb.Error
	}
	return conf.GetSizeInBytes(key), nil
}

func parseConf(c *conf, name, confType string, changeFunc func(*ViperConf), onRemoveFunc func(*ViperConf)) (*viper.Viper, error) {
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
	if changeFunc == nil {
		changeFunc = DefaultOnChangeFunc
	}
	if onRemoveFunc != nil {
		onRemoveFunc = DefaultOnRemoveFunc
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

func confChangePool(key string, c *conf) {
	v, ok := c.Val[key]
	if !ok {
		return
	}
	if v.OnChangeFunc != nil {
		v.OnChangeFunc(v)
	}
}

func confRemovePool(key string, c *conf) {
	v, ok := c.Val[key]
	if !ok {
		return
	}
	if v.OnRemoveFunc != nil {
		v.OnRemoveFunc(v)
	}
}

func DefaultOnChangeFunc(v *ViperConf) {
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
}

func DefaultOnRemoveFunc(v *ViperConf) {
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
	return utils.ConcatenateStrings(name, ".", confType)
}
