package panfigure

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

func file() error {
	configPaths := viper.Get("config_paths").([]string)
	for _, p := range configPaths {
		viper.AddConfigPath(p)
	}

	return nil
}

func UseConfigFile(name string) error {
	viper.SetConfigName(name)
	readFn := viper.ReadInConfig
	if len(viper.AllKeys()) > 0 {
		readFn = viper.MergeInConfig
	}
	if err := readFn(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}

		log.Println("config file not found: ", name)
	} else {
		file := viper.ConfigFileUsed()
		meta.filesParsed = append(meta.filesParsed, file)
		source := fmt.Sprintf("file(%s)", name)
		meta.updateSources(source)
	}

	return nil
}
