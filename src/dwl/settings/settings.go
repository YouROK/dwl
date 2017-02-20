package settings

import (
	"fmt"
	"strconv"

	"encoding/json"
)

type Settings struct {
	Url      string `json:"Url"`
	FilePath string `json:"FilePath"`

	Threads        int   `json:"Threads"`
	LoadBufferSize int64 `json:"BufferSize"`

	Config Config `json:"Config,omitempty"`
}

func NewSettings() *Settings {
	s := new(Settings)
	s.Threads = 10
	s.LoadBufferSize = 65 * 1024
	s.Config = NewConfig()
	return s
}

func FromJson(js string) (*Settings, error) {
	s := new(Settings)
	err := json.Unmarshal([]byte(js), s)
	return s, err
}

func ToJson(sets *Settings) (string, error) {
	buf, err := json.MarshalIndent(sets, "", " ")
	return string(buf), err
}

type Config map[string]interface{}

//types
// int
// int64
// string
// float64
// *Config

func NewConfig() Config {
	c := make(map[string]interface{})
	return c
}

func (c Config) Get(key string) interface{} {
	return c[key]
}

func (c Config) GetCfg(key string) Config {
	if val, ok := c[key].(Config); ok {
		return val
	}
	return nil
}

func (c Config) GetStr(key string) string {
	val, ok := c[key]
	if !ok {
		return ""
	}
	if val == nil {
		return ""
	}

	switch v := val.(type) {
	case int, int64:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	case float64, float32:
		return fmt.Sprintf("%f", v)
	}
	return ""
}

func (c Config) GetInt(key string) int {
	val, ok := c[key]
	if !ok {
		return 0
	}
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case int:
		return int(v)
	case int64:
		return int(v)
	case string:
		{
			i, _ := strconv.ParseInt(v, 10, 64)
			return int(i)
		}
	case float64:
		return int(v)
	case float32:
		return int(v)
	}
	return 0
}

func (c Config) GetInt64(key string) int64 {
	val, ok := c[key]
	if !ok {
		return 0
	}
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case int:
		return int64(v)
	case int64:
		return int64(v)
	case string:
		{
			i, _ := strconv.ParseInt(v, 10, 64)
			return int64(i)
		}
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	}
	return 0
}

func (c Config) GetFlt(key string) float64 {
	val, ok := c[key]
	if !ok {
		return 0
	}
	if val == nil {
		return 0
	}

	switch v := val.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		{
			f, _ := strconv.ParseFloat(v, 64)
			return f
		}
	case float64:
		return v
	case float32:
		return float64(v)
	}
	return 0
}

func (c Config) Set(key string, val interface{}) Config {
	c[key] = val
	return c
}
