package dbs

import "time"

type Config struct {
	Default struct {
		DB     string `yaml:"db"`
		Prefix string `yaml:"prefix"`
	} `yaml:"default"`

	DBs map[string]*DBConfig `yaml:"dbs"`
}

type DBConfig struct {
	Driver      string `yaml:"driver"`
	Dsn         string `yaml:"dsn"`
	Prefix      string `yaml:"prefix,omitempty"`
	Connections struct {
		Pool         int           `yaml:"pool"`
		Max          int           `yaml:"max"`
		Life         string        `yaml:"life"`
		LifeDuration time.Duration `yaml:",omitempty"`
	} `yaml:"connections"`

	Models struct {
		Package string `yaml:"package"`
	} `yaml:"models"`
}
