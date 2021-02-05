// преобразование типов для чтения конфигурации из файла
package config

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/buildboxapp/lib"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/gommon/color"
	"os"
	"strings"
)

var warning = color.Red("[Fail]")

// читаем конфигурации
func (c *Config) Load(configfile string) (err error) {
	fileName := ""
	cfgfile := ""

	rootDir, err := lib.RootDir()
	startDir := rootDir + sep + "upload"

	// временно, пока не перешли полностью на cfg (позже удалить)
	if len(configfile) > 5 {
		if configfile[len(configfile)-5:] != ".json" {
			configfile = configfile + ".json"
		}
	}

	if fileName, err = lib.ReadConfAction(startDir, configfile,false); err != nil {
		fmt.Printf("%s Error load enviroment: %s (configfile: %s)\n", warning, err, configfile)
		os.Exit(1)
	}

	if len(fileName) > 5 {
		if fileName[len(fileName)-5:] == ".json" {
			cfgfile = fileName[:len(fileName)-5]
		}
		cfgfile = cfgfile + ".cfg"
	}
	if _, err = toml.DecodeFile(cfgfile, &c); err != nil {
		fmt.Printf("%s Error: %s (configfile: %s)\n", warning, err, cfgfile)
		os.Exit(1)
	}

	return err
}

// формируем ClientPath из Domain
func (c *Config) SetClientPath()  {
	pp := strings.Split(c.Domain, "/")
	name := "buildbox"
	version := "gui"

	if len(pp) == 1 {
		name = pp[0]
	}
	if len(pp) == 2 {
		name = pp[0]
		version = pp[1]
	}
	c.ClientPath = "/" + name + "/" + version

	return
}

// получаем название конфигурации по-умолчанию (стоит галочка=ON)
func (c *Config) SetConfigName()  {
	fileconfig, err := lib.DefaultConfig()
	if err != nil {
		return
	}
	c.ConfigName = fileconfig
}

// задаем директорию по-умолчанию
func (c *Config) SetRootDir()  {
	rootdir, err := lib.RootDir()
	if err != nil {
		return
	}
	c.RootDir = rootdir
}


// получаем значение из конфигурации по ключу
func (c *Config) GetValue(key string) (result string, err error) {
	var rr = map[string]interface{}{}
	var flagOk = false

	// преобразуем значение типа конфигурации в структуру и получем значения в тексте
	b1, _ := json.Marshal(c)
	json.Unmarshal(b1, &rr)

	for i, v := range rr {
		if i == key {
			result = fmt.Sprint(v)
			flagOk = true
		}
	}
	if !flagOk {
		err = fmt.Errorf("%s", "Value from key not found")
	}
	return
}

// инициируем переменную значениями по-умолчанию (из структуры с дефалтовыми значениями)
func New(configfile string) Config {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		fmt.Printf("%s Error load default enviroment: %s\n", warning, err)
		os.Exit(1)
	}

	cfg.Load(configfile)
	cfg.UidService = strings.Split(configfile, ".")[0]

	return cfg
}