package logic

import (
	"github.com/adrg/xdg"
)

type Config struct {
	DownloadPath        string
	ConcurrentDownloads int
	MaximumQuality      Quality
	OutputFilePattern   string
}

func NewConfig() *Config {
	return &Config{
		DownloadPath:        xdg.UserDirs.Download,
		ConcurrentDownloads: 2,
		MaximumQuality:      QualityBest,
		OutputFilePattern:   "South_Park_$L_$S_$E_$Q_$T",
	}
}
