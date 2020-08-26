# viper_conf
基于

## easy use viper

if you have a config file is "/data/conf/debug/test.json", you can use like this.

file: /data/conf/debug/test.json
```json
{
   "abc": "abc"
}
```

core file : only New Once

```golang


// it will search config in /data/conf/debug first. 
// if not found , it will search config in /data/conf/
var Conf = NewConf("debug", "/data/conf")

```

used file : instance viper
```golang
// should use tag:mapstructure to Marshal
type AbcStruct struct {
	Abc string `json:"abc" mapstructure:"abc"`
}


abc := &AbcStruct{}
viperConf, _ := core.Conf.Instance("test.json", abc)

// get value struct
viperConf.GetValue().(*AbcStruct)

// get value like use viper .
viperConf.GetString("abc")
```

you can define self onchangefunc

```golang
func MyOnchangeFunc(v *viper_conf.ViperConf) {
    // you method
}

viperConf, _ := core.Conf.Instance("test.json", abc, viper_conf.OptOnChangeFunc(MyOnchangeFunc))

```
