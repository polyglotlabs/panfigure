package panfigure

import (
	"fmt"

	"github.com/spf13/viper"
)

// does this need to be exported?
func File() error {
	viper.SetConfigType("json")
	configPaths := viper.Get("config_paths").([]string)
	for _, p := range configPaths {
		viper.AddConfigPath(p)
	}

	// app managed (generated)
	useConfig("managed.pando")

	// user managed
	useConfig("pando")

	return nil
}

func useConfig(name string) error {
	viper.SetConfigName(name)
	readFn := viper.ReadInConfig
	if len(viper.AllKeys()) > 0 {
		readFn = viper.MergeInConfig
	}
	if err := readFn(); err != nil {
		// TODO log debug no config found with name
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	} else {
		file := viper.ConfigFileUsed()
		meta.filesParsed = append(meta.filesParsed, file)
		source := fmt.Sprintf("file(%s)", name)
		meta.updateSources(source)
	}

	return nil
}
