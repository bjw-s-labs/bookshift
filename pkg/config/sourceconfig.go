package config

import "github.com/go-playground/sensitive"

type NfsNetworkShareConfig struct {
	Host                     string `yaml:"host" validate:"required"`
	Port                     int    `yaml:"port"`
	Folder                   string `yaml:"folder" validate:"required"`
	KeepFolderStructure      bool   `yaml:"keep_folderstructure"`
	RemoveFilesAfterDownload bool   `yaml:"remove_files_after_download"`
}

type SmbNetworkShareConfig struct {
	Host                     string           `yaml:"host" validate:"required"`
	Port                     int              `yaml:"port"`
	Username                 string           `yaml:"username"`
	Password                 sensitive.String `yaml:"password"`
	Domain                   string           `yaml:"domain"`
	Share                    string           `yaml:"share" validate:"required"`
	Folder                   string           `yaml:"folder" validate:"required"`
	KeepFolderStructure      bool             `yaml:"keep_folderstructure"`
	RemoveFilesAfterDownload bool             `yaml:"remove_files_after_download"`
}
