package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// Config 结构定义了配置文件的结构
type Config struct {
	//URLs1       []string `yaml:"urls1"`
	//URLs2       []string `yaml:"urls2"`
	//URLs3       []string `yaml:"urls3"`

	ClientRequestUrl        []string `yaml:"clientRequestUrl"`
	ClientRequestSpecialUrl []string `yaml:"clientRequestSpecialUrl"`

	JobRequestUrl        []string `yaml:"jobRequestUrl"`
	JobRequestSpecialUrl []string `yaml:"jobRequestSpecialUrl"`

	SpecialMethods   []string `yaml:"specialMethods"`
	SpecialMethodMap map[string]struct{}

	LogPath            string `yaml:"logPath"`
	LoopSeconds        int    `yaml:"loopSeconds"`
	CacheSeconds       int    `yaml:"cacheSeconds"`
	BackendHttpSeconds int    `yaml:"backendHttpSeconds"`
	FrontHttpSeconds   int    `yaml:"frontHttpSeconds"`
	ChannelSize        int    `yaml:"channelSize"`
	PullJobSize        int    `yaml:"pullJobSize"`
	Chain              int    `yaml:"chain"`
	Port               int    `yaml:"port"`
}

// LoadConfig 从指定路径加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	byteValue, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(byteValue, &config); err != nil {
		return nil, err
	}

	config.SpecialMethodMap = make(map[string]struct{})
	for _, method := range config.SpecialMethods {
		config.SpecialMethodMap[method] = struct{}{}
	}

	return &config, nil
}
