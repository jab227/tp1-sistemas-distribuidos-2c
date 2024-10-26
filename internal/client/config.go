package client

import "github.com/jab227/tp1-sistemas-distribuidos-2c/internal/cprotocol"

type Config struct {
	ServerName    string            `json:"ServerName"`
	ServerPort    int               `json:"ServerPort"`
	GamesConfig   FileBatcherConfig `json:"GamesConfig"`
	ReviewsConfig FileBatcherConfig `json:"ReviewsConfig"`
	LoggerLevel   string            `json:"LoggerLevel"`
}

type FileBatcherConfig struct {
	FileType       cprotocol.FileType `json:"FileType"`
	Path           string             `json:"Path"`
	NLinesFromDisk int                `json:"NLinesFromDisk"`
	MaxElements    int                `json:"MaxElements"`
	MaxBytes       int                `json:"MaxBytes"`
}
