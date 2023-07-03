package logic

import (
	"context"

	sp "github.com/xypwn/southpark-downloader-ui/pkg/southpark"
)

type Season struct {
	sp.Season
	Region   sp.RegionInfo
	Index    int
	Episodes []sp.Episode
	MGID     string
}

func GetSeason(ctx context.Context, season sp.Season) (episodes []sp.Episode, mgid string, err error) {
	return sp.GetEpisodes(ctx, season)
}

type Series struct {
	Region  sp.RegionInfo
	Seasons map[sp.Language][]Season
	MGID    string
}

// Gets current region info and season metadata.
func GetSeries(ctx context.Context) (region sp.RegionInfo, seasons map[sp.Language][]Season, mgid string, err error) {
	region, err = sp.GetRegionInfo(ctx)
	if err != nil {
		return sp.RegionInfo{}, nil, "", err
	}

	seasons = make(map[sp.Language][]Season)

	for _, language := range region.AvailableLanguages {
		var seasonsArr []sp.Season
		seasonsArr, mgid, err = sp.GetSeasons(ctx, region, language)
		if err != nil {
			return sp.RegionInfo{}, nil, "", err
		}

		var seasonsConv []Season
		for i, v := range seasonsArr {
			seasonsConv = append(seasonsConv, Season{
				Season: v,
				Region: region,
				Index:  i,
			})
		}

		seasons[language] = seasonsConv
	}

	return
}

type Cache struct {
	Series map[sp.Host]Series
}

func NewCache() *Cache {
	return &Cache{
		Series: make(map[sp.Host]Series),
	}
}
