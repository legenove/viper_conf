# viper_conf


## easy use viper

if you have a config file is "/data/conf/debug/test.json", you can use like this.

file: /data/conf/debug/test.json
```json
{
   "abc": "abc"
}
```

```golang

// it will search config in /data/conf/debug first. 
// if not found , it will search config in /data/conf/
var Conf = NewConf("debug", "/data/conf")

abc := &AbcStruct{}
Conf.Instance("test.json", abc, nil, nil)
```
