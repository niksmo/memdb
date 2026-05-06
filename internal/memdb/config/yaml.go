package config

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

const defaultPath = "memdb.conf.yml"

type yamlProvider struct {
	args     []string
	rootPath string
}

func (fp *yamlProvider) Bytes() (yaml []byte, err error) {
	configPath := fp.parseConfigPath()

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = file.Close()
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (fp *yamlProvider) parseConfigPath() string {
	flagSet := pflag.NewFlagSet("memdb", pflag.ExitOnError)

	configPathPtr := flagSet.StringP("config", "c", defaultPath, "set the configuration filepath")

	_ = flagSet.Parse(fp.args[1:])

	configPath := *configPathPtr

	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(fp.rootPath, configPath)
	}

	return configPath
}
