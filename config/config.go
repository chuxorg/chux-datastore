// Package: config
// Description: This file contains the code to load the configuration file
// and unmarshal it into a struct.
// Author: Chuck Sailer
// Date: 2023-04-07
// Version: 1.0.0
// License: gpl-3.0
// Copyright: Chuck Sailer
// Credits: [Chuck Sailer]
// Maintainer: Chuck Sailer
// Status: Production
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// DataStoreConfig is the struct that holds the configuration
// for the data store.
// The struct is populated by the configuration file.
// The struct is used to initialize the data store.
// The struct is also used to initialize the logger.
//
// The configuration file is in YAML format.
//
// The DataStoreConfig is provided so chux-datastore can be used
// as a a stand-alone library. However, the typical use case is
// to use chux-datastore as a library in a larger application.
// In that case, the larger application will have its own
// configuration file. The larger application will use the
// chux-datastore library to initialize the data store.
// The larger application will also use the chux-datastore
// library to initialize the logger.
//
// In these cases, the larger application will have its own
// configuration file and will pass the configuration section
// to chux-datastore.
type DataStoreConfig struct {
	Logging struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logging"`
	DataStores []struct {
		DataStore struct {
			Mongo struct {
				URI            string `mapstructure:"uri"`
				Timeout        int    `mapstructure:"timeout"`
				DatabaseName   string `mapstructure:"databaseName"`
				CollectionName string `mapstructure:"collectionName"`
			} `mapstructure:"mongo"`
			Redis struct {
				URI string `mapstructure:"uri"`
			} `mapstructure:"redis"`
		} `mapstructure:"dataStore"`
	} `mapstructure:"DataStores"`
}

func New(options ...func(*DataStoreConfig)) *DataStoreConfig {

	dsc := &DataStoreConfig{}
	for _, o := range options {
		o(dsc)
	}
	return dsc 
}

func WithDataStoreConfig(dsc *DataStoreConfig) func(*DataStoreConfig) {
	return func(cfg *DataStoreConfig) {
		cfg = dsc
	}
}


func LoadConfig(configPath string) (*DataStoreConfig, error) {
	v := viper.New()

	// Set the configuration file format and name
	v.SetConfigType("yaml")
	v.SetConfigName("config")

	// Add the configuration file path
	v.AddConfigPath(configPath)

	// Read the configuration file
	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read the configuration file: %w", err)
	}

	// Unmarshal the configuration into the Config struct
	var cfg DataStoreConfig
	err = v.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the configuration: %w", err)
	}

	return &cfg, nil
}
