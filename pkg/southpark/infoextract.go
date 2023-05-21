package southpark

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/xypwn/southpark-downloader-ui/pkg/httputils"
)

func getSPBaseURL(fullURL string) (string, error) {
	url, err := url.Parse(fullURL)
	if err != nil {
		return "", err
	}
	url.Path = ""
	url.RawQuery = ""
	url.Fragment = ""

	return url.String(), nil
}

type Language int

const (
	LanguageEnglish Language = iota
	LanguageGerman
)

func LanguageFromString(s string) (lang Language, ok bool) {
	switch s {
	case LanguageEnglish.String():
		return LanguageEnglish, true
	case LanguageGerman.String():
		return LanguageGerman, true
	default:
		return 0, false
	}
}

func (l Language) String() string {
	switch l {
	case LanguageEnglish:
		return "English"
	case LanguageGerman:
		return "German"
	default:
		panic("Language.String called on invalid language")
	}
}

type RegionInfo struct {
	Host               string
	AvailableLanguages []Language
	RequiresExplicitEN bool
}

func GetRegionInfo(ctx context.Context) (RegionInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://southparkstudios.com", nil)
	if err != nil {
		return RegionInfo{}, fmt.Errorf("create southpark website request: %w", err)
	}
	var redirHost string
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirHost = req.URL.Host
			return nil
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return RegionInfo{}, fmt.Errorf("get southpark website: %w", err)
	}
	resp.Body.Close()

	res := RegionInfo{
		Host:               redirHost,
		AvailableLanguages: []Language{LanguageEnglish},
	}
	redirHost = strings.TrimPrefix(redirHost, "www.")
	switch redirHost {
	case "southpark.de":
		res.AvailableLanguages = append(res.AvailableLanguages, LanguageGerman)
		res.RequiresExplicitEN = true
		return res, nil
	case "southparkstudios.com", "southparkstudios.nu", "southparkstudios.dk", "southpark.cc.com", "southpark.nl":
		return res, nil
	default:
		return RegionInfo{}, fmt.Errorf("unsupported website region: %v", redirHost)
	}
}

func (r RegionInfo) GetURLLanguage(spURL string) (Language, error) {
	if len(r.AvailableLanguages) == 1 {
		return r.AvailableLanguages[0], nil
	} else if len(r.AvailableLanguages) == 2 &&
		r.RequiresExplicitEN {
		url, err := url.Parse(spURL)
		if err != nil {
			return 0, fmt.Errorf("parse URL: %w", err)
		}
		if strings.HasPrefix(url.Path, "/en/") {
			return LanguageEnglish, nil
		} else {
			for _, v := range r.AvailableLanguages {
				if v != LanguageEnglish {
					return v, nil
				}
			}
		}
	}
	panic("RegionInfo.GetURLLanguage called with invalid RegionInfo or invalid URL")
}

type websiteDataProps struct {
	Type    string `json:"type"`
	Filters struct {
		Items []struct {
			Label string `json:"label"`
			URL   string `json:"url"`
		} `json:"items"`
		SelectedIndex int `json:"selectedIndex"`
	} `json:"filters"`
	IsEpisodes bool `json:"isEpisodes"`
	Items      []struct {
		Label        string `json:"label"`
		URL          string `json:"url"`
		SeasonNumber int    `json:"seasonNumber"`
		Media        struct {
			Image struct {
				URL string `json:"url"`
			} `json:"image"`
			LockedLabel string `json:"lockedLabel"`
		} `json:"media"`
		Meta struct {
			SubHeader   string `json:"subHeader"`
			Description string `json:"description"`
			ItemMGID    string `json:"itemMgid"`
			SeriesMGID  string `json:"seriesMgid"`
			SeasonMGID  string `json:"seasonMgid"`
		} `json:"meta"`
	} `json:"items"`
}

type websiteData struct {
	Children []struct {
		Type     string `json:"type"`
		Children []struct {
			Type  string           `json:"type"`
			Props websiteDataProps `json:"props"`
		} `json:"children"`
	} `json:"children"`
}

func getWebsiteDataPropsFromURL(ctx context.Context, url string, containerType string, propsType string) (websiteDataProps, error) {
	body, err := httputils.GetBodyWithContext(ctx, url)
	if err != nil {
		return websiteDataProps{}, err
	}

	return getWebsiteDataPropsFromBody(body, containerType, propsType)
}

func getWebsiteDataPropsFromBody(body []byte, containerType string, propsType string) (websiteDataProps, error) {
	re := regexp.MustCompile("window.__DATA__\\s*=\\s*({.*});\n")
	match := re.FindSubmatch(body)
	if match == nil || len(match) != 2 {
		return websiteDataProps{}, errors.New("unable to find JSON data in webpage")
	}
	dataJSON := match[1]

	var data websiteData
	err := json.Unmarshal(dataJSON, &data)
	if err != nil {
		return websiteDataProps{}, fmt.Errorf("parse data JSON: %w", err)
	}

	for _, v := range data.Children {
		if v.Type == "MainContainer" {
			for _, v := range v.Children {
				if v.Type == containerType &&
					(v.Props.Type == propsType || propsType == "") {
					return v.Props, nil
				}
			}
		}
	}

	return websiteDataProps{}, fmt.Errorf("unable to find container '%s' in webpage JSON", containerType)
}

type Season struct {
	SeasonNumber int // From 1
	Title        string
	URL          string
	Language     Language
}

func GetSeasons(ctx context.Context, regionInfo RegionInfo, language Language) (seasons []Season, seriesMGID string, err error) {
	langAvailable := false
	for _, v := range regionInfo.AvailableLanguages {
		if v == language {
			langAvailable = true
			break
		}
	}

	if !langAvailable {
		return nil, "", fmt.Errorf("language '%v' not available on '%v'",
			language.String(),
			regionInfo.Host)
	}

	langPath := ""
	if language == LanguageEnglish && regionInfo.RequiresExplicitEN {
		langPath = "/en"
	}
	anySeasonURL := fmt.Sprintf("https://%v%v/seasons/south-park", regionInfo.Host, langPath)

	baseURL, err := getSPBaseURL(anySeasonURL)
	if err != nil {
		return nil, "", fmt.Errorf("get base URL: %w", err)
	}

	body, err := httputils.GetBodyWithContext(ctx, anySeasonURL)

	// Retrieve series MGID
	{
		props, err := getWebsiteDataPropsFromBody(body, "LineList", "video-guide")
		if err != nil {
			return nil, "", fmt.Errorf("retrieve series MGID in website data JSON: %w", err)
		}

		if len(props.Items) == 0 {
			return nil, "", fmt.Errorf("no series MGID found in website data JSON")
		}

		seriesMGID = props.Items[0].Meta.SeriesMGID
	}

	// Retrieve raw seasons data
	props, err := getWebsiteDataPropsFromBody(body, "SeasonSelector", "")
	if err != nil {
		return nil, "", fmt.Errorf("get 'SeasonSelector' in website data JSON: %w", err)
	}

	// Transform elements into our struct and return
	var res []Season
	for _, v := range props.Items {
		var url string
		if v.URL != "" {
			url = baseURL + v.URL
		} else {
			// If v.URL is empty, that means
			// we're at our initial URL
			url = anySeasonURL
		}
		res = append(res, Season{
			SeasonNumber: v.SeasonNumber,
			Title:        v.Label,
			URL:          url,
			Language:     language,
		})
	}

	// Sort seasons
	sort.Slice(res, func(i, j int) bool {
		return res[i].SeasonNumber < res[j].SeasonNumber
	})

	return res, seriesMGID, nil
}

type RawThumbnailURL string

func (r RawThumbnailURL) GetThumbnailURL(width uint, height uint, crop bool) string {
	return fmt.Sprintf("%v&width=%v&height=%v&crop=%v", r, width, height, crop)
}

type Episode struct {
	SeasonNumber    int // From 1
	EpisodeNumber   int // From 1
	Unavailable     bool
	RawThumbnailURL RawThumbnailURL
	Title           string
	Description     string
	MGID            string
	URL             string
	Language        Language
}

func GetEpisodes(ctx context.Context, season Season) (episodes []Episode, seasonMGID string, err error) {
	baseURL, err := getSPBaseURL(season.URL)
	if err != nil {
		return nil, "", fmt.Errorf("get base URL: %w", err)
	}

	// Get the 'Show More' API call URL
	var showMoreURL string
	{
		props, err := getWebsiteDataPropsFromURL(ctx, season.URL, "LineList", "video-guide")
		if err != nil {
			return nil, "", fmt.Errorf("retrieve 'show more' URL in website data JSON: %w", err)
		}

		index := props.Filters.SelectedIndex
		if index < 0 || index >= len(props.Filters.Items) {
			return nil, "", errors.New("invalid JSON data: index out of bounds")
		}

		showMoreURL = props.Filters.Items[index].URL
	}

	// Fetch all episodes using API call
	var res []Episode
	{
		body, err := httputils.GetBodyWithContext(ctx, baseURL+showMoreURL)
		if err != nil {
			return nil, "", fmt.Errorf("get episodes: %w", err)
		}

		var props websiteDataProps
		err = json.Unmarshal(body, &props)
		if err != nil {
			return nil, "", fmt.Errorf("parse episodes from JSON: %w", err)
		}

		if len(props.Items) == 0 {
			return nil, "", fmt.Errorf("no episodes found in season")
		}

		seasonMGID = props.Items[0].Meta.SeasonMGID

		for _, v := range props.Items {
			// Probably not the best way, but the URL always ends with "-seasonNum-XX-ep-YY",
			// so we just get the split separated by "-" as the episode number
			sp := strings.Split(v.URL, "-")
			if len(sp) < 1 {
				return nil, "", errors.New("invalid episode URL: unable to find episode number")
			}
			episodeNum, err := strconv.ParseInt(sp[len(sp)-1], 10, 32)
			if err != nil {
				return nil, "", fmt.Errorf("invalid episode URL: unable to parse episode number: %w", err)
			}
			res = append(res, Episode{
				SeasonNumber:    season.SeasonNumber,
				EpisodeNumber:   int(episodeNum),
				Unavailable:     v.Media.LockedLabel != "",
				RawThumbnailURL: RawThumbnailURL(v.Media.Image.URL),
				Title:           v.Meta.SubHeader,
				Description:     v.Meta.Description,
				MGID:            v.Meta.ItemMGID,
				URL:             baseURL + v.URL,
				Language:        season.Language,
			})
		}
	}

	// Sort episodes
	sort.Slice(res, func(i, j int) bool {
		return res[i].EpisodeNumber < res[j].EpisodeNumber
	})

	return res, seasonMGID, nil
}

type searchData struct {
	Response struct {
		Items []struct {
			Media struct {
				Image struct {
					URL string `json:"url"`
				} `json:"image"`
				LockedLabel string `json:"lockedLabel"`
			} `json:"media"`
			Meta struct {
				SubHeader   string `json:"subHeader"`
				Description string `json:"description"`
			} `json:"meta"`
			URL string `json:"url"`
		} `json:"items"`
	} `json:"response"`
}

type SearchResult struct {
	Language        Language
	Unavailable     bool
	RawThumbnailURL RawThumbnailURL
	Title           string
	Description     string
	URL             string
}

func Search(
	ctx context.Context,
	regionInfo RegionInfo,
	seriesMGID string,
	query string,
	pageNumber int, // From 0
	resultsPerPage int,
) ([]SearchResult, error) {
	showID, ok := cutPrefix(seriesMGID, "mgid:arc:series:southpark.intl:")
	if !ok {
		return nil, fmt.Errorf("invalid series MGID: %v", seriesMGID)
	}

	apiURL := fmt.Sprintf(
		"https://%v/api/search?q=%v&activeTab=Episode&showId=%v&pageNumber=%v&rowsPerPage=%v",
		regionInfo.Host,
		url.QueryEscape(query),
		showID,
		pageNumber,
		resultsPerPage,
	)

	body, err := httputils.GetBodyWithContext(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("make search API call: %w", err)
	}

	var data searchData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("parse search API response: %w", err)
	}

	var res []SearchResult
	for _, v := range data.Response.Items {
		urlLang, err := regionInfo.GetURLLanguage(v.URL)
		if err != nil {
			return nil, fmt.Errorf("get search result language: %w", err)
		}
		res = append(res, SearchResult{
			Language:        urlLang,
			Unavailable:     v.Media.LockedLabel != "",
			RawThumbnailURL: RawThumbnailURL("https:" + v.Media.Image.URL),
			Title:           v.Meta.SubHeader,
			Description:     v.Meta.Description,
			URL:             v.URL,
		})
	}
	return res, nil
}
