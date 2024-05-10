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

// Returned season and episode numbers are 1-indexed
func getSeasonAndEpisodeNumberFromURL(episodeURL string) (int, int, error) {
	sp := strings.Split(episodeURL, "-")
	if len(sp) < 3 {
		return 0, 0, fmt.Errorf("invalid URL: unable to find season and episode number: %v", episodeURL)
	}

	season, err := strconv.Atoi(sp[len(sp)-3])
	if err != nil {
		return 0, 0, fmt.Errorf("parsing season number: %w", err)
	}
	episode, err := strconv.Atoi(sp[len(sp)-1])
	if err != nil {
		return 0, 0, fmt.Errorf("parsing episode number: %w", err)
	}

	return season, episode, nil
}

type Language int

const (
	LanguageEnglish Language = iota
	LanguageGerman
)

func LanguageFromString(s string) (lang Language, ok bool) {
	switch strings.ToUpper(s) {
	case "EN", "ENGLISH":
		return LanguageEnglish, true
	case "DE", "GERMAN", "DEUTSCH":
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

type Host int

const (
	HostSPDE    Host = iota // southpark.de
	HostSPSCOM              // southparkstudios.com
	HostSPSNU               // southparkstudios.nu
	HostSPSDK               // southparkstudios.dk
	HostSPCCCOM             // southpark.cc.com
	HostSPNL                // southpark.nl
)

func HostFromString(hostStr string) (host Host, ok bool) {
	switch strings.TrimPrefix(hostStr, "www.") {
	case "southpark.de":
		return HostSPDE, true
	case "southparkstudios.com":
		return HostSPSCOM, true
	case "southparkstudios.nu":
		return HostSPSNU, true
	case "southparkstudios.dk":
		return HostSPSDK, true
	case "southpark.cc.com":
		return HostSPCCCOM, true
	case "southpark.nl":
		return HostSPNL, true
	default:
		return 0, false
	}
}

func (h Host) String() string {
	switch h {
	case HostSPDE:
		return "www.southpark.de"
	case HostSPSCOM:
		return "www.southparkstudios.com"
	case HostSPSNU:
		return "www.southparkstudios.nu"
	case HostSPSDK:
		return "www.southparkstudios.dk"
	case HostSPCCCOM:
		return "southpark.cc.com"
	case HostSPNL:
		return "www.southpark.nl"
	default:
		panic("Host.String called on invalid host")
	}
}

type RegionInfo struct {
	Host               Host
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

	if host, ok := HostFromString(redirHost); ok {
		res := RegionInfo{
			Host:               host,
			AvailableLanguages: []Language{LanguageEnglish},
		}
		if host == HostSPDE {
			res.AvailableLanguages = append(res.AvailableLanguages, LanguageGerman)
			res.RequiresExplicitEN = true
			return res, nil
		} else {
			return res, nil
		}
	} else {
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
	Media struct {
		Image struct {
			URL string `json:"url"`
		} `json:"image"`
		LockedLabel string `json:"lockedLabel"`
		Video       struct {
			Config struct {
				URI   string `json:"uri"`
				Title string `json:"title"`
			} `json:"config"`
		} `json:"video"`
		UnavailableSlate struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"unavailableSlate"`
	} `json:"media"`
	SocialShare struct {
		ShortURL string `json:"shortUrl"`
	} `json:"socialShare"`
	Meta struct {
		Description string `json:"description"`
	} `json:"meta"`
}

type websiteData struct {
	Children []struct {
		Type     string `json:"type"`
		Children []struct {
			Type  string           `json:"type"`
			Props websiteDataProps `json:"props"`
		} `json:"children"`
		HandleTVEAuthRedirection *struct {
			VideoDetail struct {
				VideoServiceURL string `json:"videoServiceUrl"`
			} `json:"videoDetail"`
		} `json:"handleTVEAuthRedirection"`
	} `json:"children"`
}

func getWebsiteDataFromURL(ctx context.Context, url string) (websiteData, error) {
	body, err := httputils.GetBodyWithContext(ctx, url)
	if err != nil {
		return websiteData{}, err
	}

	return getWebsiteDataFromBody(body)
}

func getWebsiteDataFromBody(body []byte) (websiteData, error) {
	re := regexp.MustCompile("window.__DATA__\\s*=\\s*({.*});\n")
	match := re.FindSubmatch(body)
	if match == nil || len(match) != 2 {
		return websiteData{}, errors.New("unable to find JSON data in webpage")
	}
	dataJSON := match[1]

	var data websiteData
	err := json.Unmarshal(dataJSON, &data)
	if err != nil {
		return websiteData{}, fmt.Errorf("parse data JSON: %w", err)
	}
	return data, nil
}

func getWebsiteDataPropsFromURL(ctx context.Context, url string, containerType string, propsType string) (websiteDataProps, error) {
	body, err := httputils.GetBodyWithContext(ctx, url)
	if err != nil {
		return websiteDataProps{}, err
	}

	return getWebsiteDataPropsFromBody(body, containerType, propsType)
}

func getWebsiteDataPropsFromBody(body []byte, containerType string, propsType string) (websiteDataProps, error) {
	data, err := getWebsiteDataFromBody(body)
	if err != nil {
		return websiteDataProps{}, err
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
	// Using season one instead of /seasons/south-park, because in some
	// regions (e.g. sweden), the last season isn't available, meaning we
	// can't get the series MGID via that season, which messes everything up
	anySeasonURL := fmt.Sprintf("https://%v%v/seasons/south-park/yjy8n9/season-1", regionInfo.Host, langPath)

	baseURL, err := getSPBaseURL(anySeasonURL)
	if err != nil {
		return nil, "", fmt.Errorf("get base URL: %w", err)
	}

	body, err := httputils.GetBodyWithContext(ctx, anySeasonURL)
	if err != nil {
		return nil, "", fmt.Errorf("get base data: %w", err)
	}

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

type EpisodeMetadata struct {
	SeasonNumber    int // From 1
	EpisodeNumber   int // From 1
	Language        Language
	Unavailable     bool
	RawThumbnailURL string
	Title           string
	Description     string
	URL             string
}

func (em EpisodeMetadata) GetThumbnailURL(width uint, height uint, crop bool) string {
	return fmt.Sprintf("%v&width=%v&height=%v&crop=%v", em.RawThumbnailURL, width, height, crop)
}

// Compares season, episode and language.
// NOT an exact equality check.
func (em EpisodeMetadata) Is(other EpisodeMetadata) bool {
	return em.SeasonNumber == other.SeasonNumber &&
		em.EpisodeNumber == other.EpisodeNumber &&
		em.Language == other.Language
}

type Episode struct {
	EpisodeMetadata
	MGID string
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
			seasonNum, episodeNum, err := getSeasonAndEpisodeNumberFromURL(v.URL)
			if err != nil {
				return nil, "", fmt.Errorf("extract season and episode number from URL: %w", err)
			}
			if seasonNum != season.SeasonNumber {
				return nil, "", fmt.Errorf("mismatch between season number in url (%v) and in season parameter (%v)", seasonNum, season.SeasonNumber)
			}
			res = append(res, Episode{
				EpisodeMetadata: EpisodeMetadata{
					SeasonNumber:    seasonNum,
					EpisodeNumber:   episodeNum,
					Language:        season.Language,
					Unavailable:     v.Media.LockedLabel != "",
					RawThumbnailURL: v.Media.Image.URL,
					Title:           v.Meta.SubHeader,
					Description:     v.Meta.Description,
					URL:             baseURL + v.URL,
				},
				MGID: v.Meta.ItemMGID,
			})
		}
	}

	// Sort episodes
	sort.Slice(res, func(i, j int) bool {
		return res[i].EpisodeNumber < res[j].EpisodeNumber
	})

	return res, seasonMGID, nil
}

func GetEpisode(ctx context.Context, regionInfo RegionInfo, url string) (Episode, error) {
	props, err := getWebsiteDataPropsFromURL(ctx, url, "VideoPlayer", "")
	if err != nil {
		return Episode{}, fmt.Errorf("get website data props: %w", err)
	}

	shareURL := props.SocialShare.ShortURL

	language, err := regionInfo.GetURLLanguage(shareURL)
	if err != nil {
		return Episode{}, fmt.Errorf("get episode language: %w", err)
	}

	seasonNum, episodeNum, err := getSeasonAndEpisodeNumberFromURL(shareURL)
	if err != nil {
		return Episode{}, fmt.Errorf("extract season and episode number from URL: %w", err)
	}

	return Episode{
		EpisodeMetadata: EpisodeMetadata{
			SeasonNumber:    seasonNum,
			EpisodeNumber:   episodeNum,
			Language:        language,
			Unavailable:     props.Media.UnavailableSlate.Title != "",
			RawThumbnailURL: props.Media.Image.URL,
			Title:           props.Media.Video.Config.Title,
			Description:     props.Meta.Description,
			URL:             shareURL,
		},
		MGID: props.Media.Video.Config.URI,
	}, nil
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

func Search(
	ctx context.Context,
	regionInfo RegionInfo,
	seriesMGID string,
	query string,
	pageNumber int, // From 0
	resultsPerPage int,
) ([]EpisodeMetadata, error) {
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

	var res []EpisodeMetadata
	for _, v := range data.Response.Items {
		// HACK: The "1%" (S15E12) episode's URL isn't escaped correctly. This
		// leads to a 400 error when using the official website, but correctly
		// URL escaping it seems to fix the issue.
		v.URL = strings.ReplaceAll(v.URL, "south-park-1%", "south-park-1%37")

		seasonNum, episodeNum, err := getSeasonAndEpisodeNumberFromURL(v.URL)
		if err != nil {
			return nil, fmt.Errorf("extract season and episode number from URL: %w", err)
		}

		urlLang, err := regionInfo.GetURLLanguage(v.URL)
		if err != nil {
			return nil, fmt.Errorf("get search result language: %w", err)
		}
		res = append(res, EpisodeMetadata{
			SeasonNumber:    seasonNum,
			EpisodeNumber:   episodeNum,
			Language:        urlLang,
			Unavailable:     v.Media.LockedLabel != "",
			RawThumbnailURL: "https:" + v.Media.Image.URL,
			Title:           v.Meta.SubHeader,
			Description:     v.Meta.Description,
			URL:             v.URL,
		})
	}
	return res, nil
}
