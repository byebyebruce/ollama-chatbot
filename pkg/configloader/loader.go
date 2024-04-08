package configloader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

func Load(path string, cfg interface{}) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	suffix := filepath.Ext(path)
	switch suffix {
	case ".json":
		err = json.Unmarshal(b, cfg)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(b, cfg)
	case ".toml":
		err = toml.Unmarshal(b, cfg)
	default:
		err = fmt.Errorf("unknown config file suffix %s", suffix)
	}
	return err
}
