package logic

import (
	//"github.com/xypwn/southpark-downloader-ui/pkg/data"

	"github.com/adrg/xdg"
)

/*type Config struct {
	DownloadPath data.Binding[string]
	ConcurrentDownloads data.Binding[int]
	MaximumQuality data.Binding[Quality]
}*/

type Config struct {
	DownloadPath        string
	ConcurrentDownloads int
	MaximumQuality      Quality
}

func NewConfig() *Config {
	return &Config{
		DownloadPath:        xdg.UserDirs.Download,
		ConcurrentDownloads: 2,
		MaximumQuality:      QualityBest,
	}
}
