package apputils

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

func LoadConfig(target interface{}, path string) {
	confFile, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(confFile, &config); err != nil {
		log.Fatal(err)
	}

	// 🙏 https://github.com/go-yaml/yaml/issues/13#issuecomment-428952604
	b, err := yaml.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(b, target)
	if err != nil {
		log.Fatal(err)
	}
}
