# 简单解析ini配置文件



```go
// config.go

package main

import (
	"github.com/liuq000/go-lib/config/ini"
	"time"
)

type configGlobal struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	Timeout  time.Duration
}

type configTables struct {
	Field string
	Day   int
}

type Config struct {
	Global *configGlobal
	Tables map[string]configTables
}

func (conf *Config) Parse() []error {

	cfg := ini.Load("", true)

	conf.Global = &configGlobal{
		Host:     cfg.String("global", "host"),
		Port:     cfg.String("global", "port"),
		Username: cfg.String("global", "username"),
		Password: cfg.String("global", "password"),
		DBName:   cfg.String("global", "dbname"),
		Timeout:  cfg.Duration("global", "timeout"),
	}
	conf.Tables = make(map[string]configTables)
	tablesName := cfg.SectionStrings()
	for i := range tablesName {
		conf.Tables[tablesName[i]] = configTables{
			Field: cfg.String(tablesName[i], "field"),
			Day:   cfg.Int(tablesName[i], "day"),
		}
	}

	return cfg.Errors()
}
```

```go
// main.go

package main

import (
	"log"
	"os"
)

func main() {

	conf := &Config{}
	if errs := conf.Parse(); len(errs) != 0 {
		for i := range errs {
			log.Println(errs[i])
		}
		os.Exit(1)
	}

	log.Println("conf:", conf)
}
```

