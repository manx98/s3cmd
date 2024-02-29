package main

import "gopkg.in/ini.v1"

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	SkipSSL   bool
}

func loadConfig() (*MinioConfig, error) {
	load, err := ini.Load(*configPath)
	if err != nil {
		return nil, err
	}
	return &MinioConfig{
		Endpoint:  load.Section("minio").Key("endpoint").String(),
		AccessKey: load.Section("minio").Key("access_key").String(),
		SecretKey: load.Section("minio").Key("secret_key").String(),
		UseSSL:    load.Section("minio").Key("use_ssl").MustBool(false),
		SkipSSL:   load.Section("minio").Key("skip_ssl").MustBool(false),
	}, nil
}
