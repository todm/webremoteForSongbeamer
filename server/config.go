package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	Port               int
	SongbeamerAddr     string
	UseTLS             bool
	CertFile           string
	KeyFile            string
	ThumbnailQuality   int
	ThumbnailSkipReset bool
	StaticDir          string
}

var cfg *Config = nil

func ReadConfig() error {
	if cfg != nil {
		return nil
	}

	cfg = &Config{
		Port:               10024,
		SongbeamerAddr:     "127.0.0.1:10023",
		UseTLS:             false,
		CertFile:           "",
		KeyFile:            "",
		ThumbnailQuality:   50,
		ThumbnailSkipReset: false,
		StaticDir:          "./static",
	}

	path := ResolvePath("webremote.config.json")

	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			file, err := os.Create(path)
			if err != nil {
				return err
			}
			defer file.Close()

			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(cfg); err != nil {
				return err
			}
			return nil
		} else {
			return err
		}
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return err
	}
	return nil
}

func ResolvePath(file string) string {
	if file == "" || filepath.IsAbs(file) {
		return file
	}
	exe, err := os.Executable()
	if err != nil {
		exe = "./"
	}
	return filepath.Join(filepath.Dir(exe), file)
}
