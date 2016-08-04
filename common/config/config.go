package config

import (
	slog "log"
	"strconv"
	"strings"

	conf "github.com/robfig/config"
)

var defaultConfig *conf.Config
var env map[string]string

func GetConfigString(key string) (v string, err error) {
	v, err = defaultConfig.String(env["section"], key)
	return
}

func GetConfigIntSlice(key string) (v []int, err error) {
	var vs string
	vs, err = defaultConfig.String(env["section"], key)
	if err != nil {
		return
	}
	vsslice := strings.Split(vs, ",")
	var itm int64
	for _, vi := range vsslice {
		itm, err = strconv.ParseInt(vi, 10, 32)
		if err != nil {
			return
		}
		v = append(v, int(itm))
	}
	return
}

func GetConfigInt(key string) (v int, err error) {
	v, err = defaultConfig.Int(env["section"], key)
	return
}

func GetConfigInt64(key string) (v int64, err error) {
	v32, er := defaultConfig.Int(env["section"], key)
	v = int64(v32)
	err = er
	return
}

func GetConfigBool(key string) (v bool, err error) {
	v, err = defaultConfig.Bool(env["section"], key)
	return
}

func GetConfigFloat(key string) (v float64, err error) {
	v, err = defaultConfig.Float(env["section"], key)
	return
}

func loadConfigFile() error {
	configpath, _ := env["config"]
	c, err := conf.ReadDefault(configpath)
	if err != nil {
		return err
	}
	defaultConfig.Merge(c)
	return nil
}

func InitConfig(pconfig *string, psection *string) {
	var err error
	if *pconfig == "" || *psection == "" {
		slog.Fatalf("config and section not found in Env\n")
	}
	env = map[string]string{
		"config":  *pconfig,
		"section": *psection,
	}
	configpath, _ := env["config"]
	defaultConfig, err = conf.ReadDefault(configpath)
	if err != nil {
		slog.Fatalf("ReadDefault failed. err=%v\n", err)
	}

	err = loadConfigFile()
	if err != nil {
		slog.Fatalf("LoadConfigFile failed!err:=%v", err)
	}
}
