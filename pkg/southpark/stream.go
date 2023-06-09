package southpark

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/xypwn/southpark-downloader-ui/pkg/httputils"

	"github.com/yapingcat/gomedia/go-mp4"
	"github.com/yapingcat/gomedia/go-mpeg2"
)

// Processes strings like METHOD=AES-128,URI="https://.../",IV=0xDEADBEEF
func getExtM3UInfo(data string, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		panic("getExtM3UInfo: 'v' must be a non-nil pointer to a struct")
	}
	rve := rv.Elem()
	if rve.Kind() != reflect.Struct {
		panic("getExtM3UInfo: expected a *struct as 'v'")
	}

	// Separate by comma, except when in double quotes
	quoted := false
	pairs := strings.FieldsFunc(data, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && r == ','
	})

	for _, pair := range pairs {
		sp := strings.SplitN(pair, "=", 2)
		if len(sp) != 2 {
			return fmt.Errorf("invalid ExtM3U format")
		}

		key := sp[0]
		val := sp[1]

		// Remove possible double quotes around rhs string
		if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
			val = val[1:]
			val = val[:len(val)-1]
		}

		// Set struct field
		var fName string
		f := rve.FieldByNameFunc(func(s string) bool {
			res := strings.ToUpper(s) == strings.ReplaceAll(strings.ToUpper(key), "-", "")
			if res {
				fName = s
			}
			return res
		})

		if fName == "" {
			return fmt.Errorf("unable to find a fitting field for '%v'", key)
		}

		if !f.IsValid() {
			panic("getExtM3UInfo: field '" + fName + "' of 'v' is invalid")
		}

		if !f.CanSet() {
			panic("getExtM3UInfo: field '" + fName + "' of 'v' is inaccessible")
		}

		const errUnsupportedField = "getExtM3UInfo: only strings, float32s and hexadecimal []bytes are supported as struct fields"
		switch f.Kind() {
		case reflect.String:
			f.SetString(val)
		case reflect.Int:
			i, err := strconv.ParseInt(val, 10, 32)
			if err != nil {
				return fmt.Errorf("parse int data for %v: %w", fName, err)
			}
			f.SetInt(i)
		case reflect.Float32:
			v, err := strconv.ParseFloat(val, 32)
			if err != nil {
				return fmt.Errorf("parse float data for %v: %w", fName, err)
			}
			f.SetFloat(v)
		case reflect.Slice:
			if f.Type().Elem().Kind() == reflect.Uint8 {
				val = strings.TrimPrefix(val, "0x")
				decoded, err := hex.DecodeString(val)
				if err != nil {
					return fmt.Errorf("decode hex data for %v: %w", fName, err)
				}
				f.SetBytes(decoded)
			} else {
				panic(errUnsupportedField)
			}
		default:
			panic(errUnsupportedField)
		}
	}

	return nil
}

func downloadAndDecryptAES128Segment(ctx context.Context, url string, key []byte, iv []byte) ([]byte, error) {
	data, err := httputils.GetBodyWithContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("get AES128 encrypted segment: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes.NewCipher: %w", err)
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// Make sure encrypted data length is a multiple of AES block size
	if len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("encrypted data length (%v) is not a multiple of AES block size (%v)", len(data), aes.BlockSize)
	}

	// Decrypt data
	mode.CryptBlocks(data, data)

	return data, nil
}

type feedDoc struct {
	Feed struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Image       struct {
			URL string `json:"url"`
		} `json:"image"`
		Items []struct {
			AirDate     string `json:"airDate"`
			Description string `json:"description"`
			Duration    int    `json:"duration"`
			Group       struct {
				Content string `json:"content"`
			} `json:"group"`
			Title string `json:"title"`
		} `json:"items"`
	} `json:"feed"`
}

func getMediaGenURLs(ctx context.Context, mgid string, url string) ([]string, error) {
	infoURL := fmt.Sprintf("http://media.mtvnservices.com/pmt/e1/access/index.html?uri=%v&configtype=edge&ref=%v", mgid, url)

	dataJSON, err := httputils.GetBodyWithContext(ctx, infoURL)
	if err != nil {
		return nil, fmt.Errorf("get feed doc: %w", err)
	}

	var data feedDoc
	err = json.Unmarshal(dataJSON, &data)
	if err != nil {
		return nil, fmt.Errorf("parse feed doc: %w", err)
	}

	var res []string
	for _, v := range data.Feed.Items {
		url := v.Group.Content
		url = strings.Replace(url, "&device={device}", "", 1)
		url += "&acceptMethods=hls"
		url += "&format=json"
		res = append(res, url)
	}
	return res, nil
}

type mediaGenDoc struct {
	Package struct {
		Version string `json:"version"`
		Video   struct {
			Item []struct {
				OriginationDate string `json:"origination_date"`
				Rendition       []struct {
					Cdn      string `json:"cdn"`
					Method   string `json:"method"`
					Duration string `json:"duration"`
					Type     string `json:"type"`
					Src      string `json:"src"`
					Rdcount  string `json:"rdcount"`
				} `json:"rendition"`
				Transcript []struct {
					Kind        string `json:"kind"`
					Srclang     string `json:"srclang"`
					Label       string `json:"label"`
					Typographic []struct {
						Format string `json:"format"`
						Src    string `json:"src"`
					} `json:"typographic"`
				} `json:"transcript"`
			} `json:"item"`
		} `json:"video"`
	} `json:"package"`
}

type highlevelMediaCaption struct {
	Format string
	URL    string
}

type highlevelMedia struct {
	StreamMasterURL string
	StreamMethod    string
	StreamDuration  int
	StreamType      string
	CaptionLang     string
	CaptionLabel    string
	Captions        []highlevelMediaCaption
}

func getHighlevelMedia(ctx context.Context, mediaGenURL string) (highlevelMedia, error) {
	body, err := httputils.GetBodyWithContext(ctx, mediaGenURL)
	if err != nil {
		return highlevelMedia{}, fmt.Errorf("get mediagen doc: %w", err)
	}

	var doc mediaGenDoc
	err = json.Unmarshal(body, &doc)
	if err != nil {
		return highlevelMedia{}, fmt.Errorf("parse mediagen doc: %w", err)
	}

	if len(doc.Package.Video.Item) != 1 {
		return highlevelMedia{}, fmt.Errorf("mediagen JSON: expected exactly 1 video item, but found %v",
			len(doc.Package.Video.Item))
	}
	videoItem := doc.Package.Video.Item[0]
	if len(videoItem.Rendition) != 1 {
		return highlevelMedia{}, fmt.Errorf("mediagen JSON: expected exactly 1 video rendition, but found %v",
			len(videoItem.Rendition))
	}
	rendition := videoItem.Rendition[0]
	if len(videoItem.Transcript) != 1 {
		return highlevelMedia{}, fmt.Errorf("mediagen JSON: expected exactly 1 video transcript, but found %v",
			len(videoItem.Transcript))
	}
	transcript := videoItem.Transcript[0]
	duration, err := strconv.ParseInt(rendition.Duration, 10, 32)
	if err != nil {
		return highlevelMedia{}, fmt.Errorf("parsing stream duration: %w", err)
	}

	var res highlevelMedia
	res.StreamMasterURL = rendition.Src
	res.StreamMethod = rendition.Method
	res.StreamDuration = int(duration)
	res.StreamType = rendition.Type
	res.CaptionLang = transcript.Srclang
	res.CaptionLabel = transcript.Label
	for _, t := range transcript.Typographic {
		res.Captions = append(res.Captions, highlevelMediaCaption{
			Format: t.Format,
			URL:    t.Src,
		})
	}
	return res, nil
}

type HLSFormat struct {
	AverageBandwidth uint
	FrameRate        float32
	Codecs           string
	Width            uint
	Height           uint
	Bandwidth        uint
	URL              string
}

// Formats are returned sorted from best to worst
func getHLSFormats(ctx context.Context, hlsMasterURL string) ([]HLSFormat, error) {
	body, err := httputils.GetBodyWithContext(ctx, hlsMasterURL)
	if err != nil {
		return nil, fmt.Errorf("get master HLS playlist: %w", err)
	}
	lines := strings.Split(string(body), "\n")

	var format HLSFormat
	var formats []HLSFormat
	for _, line := range lines {
		if streamInfoStr, found := cutPrefix(line, "#EXT-X-STREAM-INF:"); found {
			var streamInfo struct {
				AverageBandwidth int
				FrameRate        float32
				Codecs           string
				Resolution       string
				Bandwidth        int
			}

			err := getExtM3UInfo(streamInfoStr, &streamInfo)
			if err != nil {
				return nil, fmt.Errorf("getExtM3UInfo: %w", err)
			}

			format.AverageBandwidth = uint(streamInfo.AverageBandwidth)
			format.FrameRate = streamInfo.FrameRate
			format.Codecs = streamInfo.Codecs
			{
				sp := strings.SplitN(streamInfo.Resolution, "x", 2)
				if len(sp) != 2 {
					return nil, errors.New("invalid resolution format in EXT-X-STREAM-INF")
				}
				w, err := strconv.ParseUint(sp[0], 10, 32)
				if err != nil {
					return nil, fmt.Errorf("parse resolution width: %w", err)
				}
				format.Width = uint(w)
				h, err := strconv.ParseUint(sp[1], 10, 32)
				if err != nil {
					return nil, fmt.Errorf("parse resolution height: %w", err)
				}
				format.Height = uint(h)
			}
			format.Bandwidth = uint(streamInfo.Bandwidth)
		} else if strings.HasPrefix(line, "https://") {
			format.URL = line
			formats = append(formats, format)
			format = HLSFormat{}
		}
	}

	// Sort by bandwidth (best first)
	sort.Slice(formats, func(i, j int) bool {
		return formats[j].Bandwidth < formats[i].Bandwidth
	})

	return formats, nil
}

type HLSStreamKey struct {
	Method string
	Key    []byte
	IV     []byte
}

type HLSStreamSegment struct {
	Duration float64
	URL      string
}

type HLSStream struct {
	Key      HLSStreamKey
	Segments []HLSStreamSegment
}

func getHLSStream(ctx context.Context, hlsURL string) (HLSStream, error) {
	body, err := httputils.GetBodyWithContext(ctx, hlsURL)
	if err != nil {
		return HLSStream{}, fmt.Errorf("get stream HLS playlist: %w", err)
	}
	lines := strings.Split(string(body), "\n")

	var keyInfo struct {
		Method string
		URI    string
		IV     []byte
	}

	var duration float64 = 0
	var segments []HLSStreamSegment
	var key []byte
	for _, line := range lines {
		if keyInfoStr, found := cutPrefix(line, "#EXT-X-KEY:"); found {
			// Parse key info
			err := getExtM3UInfo(keyInfoStr, &keyInfo)
			if err != nil {
				return HLSStream{}, fmt.Errorf("getExtM3UInfo: %w", err)
			}

			// Download key
			key, err = httputils.GetBodyWithContext(ctx, keyInfo.URI)
			if err != nil {
				return HLSStream{}, fmt.Errorf("get decryption key: %w", err)
			}
		} else if infoStr, found := cutPrefix(line, "#EXTINF:"); found {
			var err error
			sp := strings.Split(infoStr, ",")
			if len(sp) < 1 {
				return HLSStream{}, errors.New("no segment duration specified in #EXTINF")
			}

			duration, err = strconv.ParseFloat(sp[0], 64)
			if err != nil {
				return HLSStream{}, fmt.Errorf("parse segment duration: %w", err)
			}
		} else if strings.HasPrefix(line, "https://") {
			if duration == 0 {
				return HLSStream{}, errors.New("no segment duration found")
			}
			segments = append(segments, HLSStreamSegment{
				URL:      line,
				Duration: duration,
			})
			duration = 0
		}
	}
	return HLSStream{
		Key: HLSStreamKey{
			Method: keyInfo.Method,
			Key:    key,
			IV:     keyInfo.IV,
		},
		Segments: segments,
	}, nil
}

func ConvertTSToMP4(tsInput io.Reader, mp4Output io.WriteSeeker) error {
	muxer, err := mp4.CreateMp4Muxer(mp4Output)
	if err != nil {
		return fmt.Errorf("create mp4 muxer: %w", err)
	}

	var writeErr error

	// https://github.com/yapingcat/gomedia/blob/main/example/example_convert_ts_to_mp4.go
	hasAudio := false
	hasVideo := false
	var atid uint32 = 0
	var vtid uint32 = 0
	demuxer := mpeg2.NewTSDemuxer()
	demuxer.OnFrame = func(cid mpeg2.TS_STREAM_TYPE, frame []byte, pts uint64, dts uint64) {
		if cid == mpeg2.TS_STREAM_H264 {
			if !hasVideo {
				vtid = muxer.AddVideoTrack(mp4.MP4_CODEC_H264)
				hasVideo = true
			}
			err := muxer.Write(vtid, frame, uint64(pts), uint64(dts))
			if err != nil {
				writeErr = err
			}
		} else if cid == mpeg2.TS_STREAM_AAC {
			if !hasAudio {
				atid = muxer.AddAudioTrack(mp4.MP4_CODEC_AAC)
				hasAudio = true
			}
			err = muxer.Write(atid, frame, uint64(pts), uint64(dts))
			if err != nil {
				writeErr = err
			}
		} else if cid == mpeg2.TS_STREAM_AUDIO_MPEG1 || cid == mpeg2.TS_STREAM_AUDIO_MPEG2 {
			if !hasAudio {
				atid = muxer.AddAudioTrack(mp4.MP4_CODEC_MP3)
				hasAudio = true
			}
			err := muxer.Write(atid, frame, uint64(pts), uint64(dts))
			if err != nil {
				writeErr = err
			}
		}
	}

	if err := demuxer.Input(tsInput); err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			// File is incomplete, ignore
		} else {
			return fmt.Errorf("MPEG-TS demuxer: %w", err)
		}
	}

	muxer.WriteTrailer()

	if writeErr != nil {
		// Currently only propagates last error
		return fmt.Errorf("mp4 muxer: %w", err)
	}

	return nil
}

type EpisodePart struct {
	Stream          HLSStream
	VTTSubtitleURLs []string
}

func GetEpisodeParts(ctx context.Context, e Episode, selectFormat func([]HLSFormat) (HLSFormat, error)) ([]EpisodePart, error) {
	mediaGenURLs, err := getMediaGenURLs(ctx, e.MGID, e.URL)
	if err != nil {
		return nil, fmt.Errorf("getMediaGenURLs: %w", err)
	}

	var parts []EpisodePart
	for _, mediaGenURL := range mediaGenURLs {
		highlevelMedia, err := getHighlevelMedia(ctx, mediaGenURL)
		if err != nil {
			return nil, fmt.Errorf("getHighlevelMedia: %w", err)
		}

		if highlevelMedia.StreamMethod != "hls" {
			return nil, fmt.Errorf("expected HLS stream, but got '%v' instead", highlevelMedia.StreamMethod)
		}

		var vttSubtitleURLs []string
		for _, v := range highlevelMedia.Captions {
			if v.Format == "vtt" {
				vttSubtitleURLs = append(vttSubtitleURLs, v.URL)
			}
		}

		hlsFormats, err := getHLSFormats(ctx, highlevelMedia.StreamMasterURL)
		if err != nil {
			return nil, fmt.Errorf("getHLSFormats: %w", err)
		}

		format, err := selectFormat(hlsFormats)
		if err != nil {
			return nil, fmt.Errorf("selectFormat: %w", err)
		}

		stream, err := getHLSStream(ctx, format.URL)
		if err != nil {
			return nil, fmt.Errorf("getHLSStream: %w", err)
		}

		if stream.Key.Method != "AES-128" {
			return nil, fmt.Errorf("unable to decrypt '%v'; only AES-128 decryption is supported", stream.Key.Method)
		}

		parts = append(parts, EpisodePart{
			Stream:          stream,
			VTTSubtitleURLs: vttSubtitleURLs,
		})
	}

	return parts, nil
}

func GetPartsTotalHLSSegments(parts []EpisodePart) int {
	n := 0
	for _, v := range parts {
		n += len(v.Stream.Segments)
	}
	return n
}

func GetEpisodeTSVideo(ctx context.Context, parts []EpisodePart, startSegment int, segmentCallback func([]byte) error) error {
	segmentIndex := 0
	for _, part := range parts {
		for _, seg := range part.Stream.Segments {
			if segmentIndex >= startSegment {
				data, err := downloadAndDecryptAES128Segment(ctx, seg.URL, part.Stream.Key.Key, part.Stream.Key.IV)
				if err != nil {
					return fmt.Errorf("downloadAndDecryptAES128Segment: %w", err)
				}
				if err := segmentCallback(data); err != nil {
					return fmt.Errorf("segmentCallback: %w", err)
				}
			}
			segmentIndex++
		}
	}
	return nil
}

func GetEpisodeVTTSubtitles(ctx context.Context, parts []EpisodePart) ([]byte, error) {
	vttParts := make([][]byte, 0, len(parts))
	durations := make([]float64, 0, len(parts))
	for i, part := range parts {
		if len(part.VTTSubtitleURLs) != 1 {
			return nil, fmt.Errorf("part %v: expected exactly 1 subtitle track, but found %v", i, len(part.VTTSubtitleURLs))
		}

		vttPart, err := httputils.GetBodyWithContext(ctx, part.VTTSubtitleURLs[0])
		if err != nil {
			return nil, err
		}
		vttParts = append(vttParts, vttPart)

		var duration float64 = 0
		for _, seg := range part.Stream.Segments {
			duration += seg.Duration
		}
		durations = append(durations, duration)
	}

	res, err := mergeVTTSubtitles(vttParts, durations)
	if err != nil {
		return nil, fmt.Errorf("mergeVTTSubtitles: %w", err)
	}
	return res, nil
}
