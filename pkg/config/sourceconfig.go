package config

import "github.com/go-playground/sensitive"

type SourceConfig interface{}

type NfsNetworkShareConfig struct {
	Host                     string `yaml:"host" validate:"required"`
	Port                     int    `yaml:"port"`
	Folder                   string `yaml:"folder" validate:"required"`
	KeepFolderStructure      bool   `yaml:"keep_folderstructure"`
	RemoveFilesAfterDownload bool   `yaml:"remove_files_after_download"`
	TimeoutSeconds           int    `yaml:"timeout_seconds"`
}

type SmbNetworkShareConfig struct {
	Host                     string            `yaml:"host" validate:"required"`
	Port                     int               `yaml:"port"`
	Username                 string            `yaml:"username"`
	Password                 *sensitive.String `yaml:"password"`
	Domain                   string            `yaml:"domain"`
	Share                    string            `yaml:"share" validate:"required"`
	Folder                   string            `yaml:"folder" validate:"required"`
	KeepFolderStructure      bool              `yaml:"keep_folderstructure"`
	RemoveFilesAfterDownload bool              `yaml:"remove_files_after_download"`
	TimeoutSeconds           int               `yaml:"timeout_seconds"`
}

type ImapConfig struct {
	Host                      string            `yaml:"host" validate:"required"`
	Port                      int               `yaml:"port"`
	Username                  string            `yaml:"username"`
	Password                  *sensitive.String `yaml:"password"`
	Mailbox                   string            `yaml:"mailbox" validate:"required"`
	FilterField               string            `yaml:"filter_field" validate:"required,oneof=to subject"`
	FilterValue               string            `yaml:"filter_value" validate:"required"`
	ProcessReadEmails         bool              `yaml:"process_read_emails"`
	RemoveEmailsAfterDownload bool              `yaml:"remove_emails_after_download"`
	TimeoutSeconds            int               `yaml:"timeout_seconds"`
}
