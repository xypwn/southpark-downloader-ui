package southpark

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/xypwn/southpark-downloader-ui/pkg/httputils"
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
			continue
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

func downloadAndDecryptAES128Segment(ctx context.Context, url string, key []byte, segmentIdx int) ([]byte, error) {
	data, err := httputils.GetBodyWithContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("get AES128 encrypted segment: %w", err)
	}

	if len(data) < 16 {
		return nil, fmt.Errorf("cipher data too short")
	}
	var iv [16]byte
	{
		// https://github.com/FFmpeg/FFmpeg/blob/d7924a4f60f2088de1e6790345caba929eb97030/libavformat/hls.c#L882
		startSeq := 1 // #EXT-X-MEDIA-SEQUENCE
		binary.BigEndian.PutUint64(iv[8:], uint64(startSeq+segmentIdx))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes.NewCipher: %w", err)
	}

	mode := cipher.NewCBCDecrypter(block, iv[:])

	// Make sure encrypted data length is a multiple of AES block size
	if len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("encrypted data length (%v) is not a multiple of AES block size (%v)", len(data), aes.BlockSize)
	}

	// Decrypt data
	mode.CryptBlocks(data, data)

	return data, nil
}

type videoServiceDoc struct {
	Stitchedstream struct {
		Source string `json:"source"`
	} `json:"stitchedstream"`
	Content []struct {
		ID       string `json:"id"`
		Chapters []struct {
			Sequence int
			ID       string `json:"id"`
		} `json:"chapters"`
	} `json:"content"`
}

func getMediaMasterURL(ctx context.Context, url string) (string, error) {
	var videoServiceURL string
	{
		data, err := getWebsiteDataFromURL(ctx, url)
		if err != nil {
			return "", fmt.Errorf("get episode website data: %w", err)
		}
		for _, v := range data.Children {
			if v.HandleTVEAuthRedirection != nil {
				videoServiceURL = v.HandleTVEAuthRedirection.VideoDetail.VideoServiceURL
				break
			}
		}
		if cutURL, _, found := strings.Cut(videoServiceURL, "?"); found {
			videoServiceURL = cutURL
		} else {
			return "", fmt.Errorf("video service URL does not contain a query: '%v'", videoServiceURL)
		}
		videoServiceURL += "?clientPlatform=desktop"
	}

	dataJSON, err := httputils.GetBodyWithContext(ctx, videoServiceURL)
	if err != nil {
		return "", fmt.Errorf("get video service doc: %w", err)
	}

	var data videoServiceDoc
	err = json.Unmarshal(dataJSON, &data)
	if err != nil {
		return "", fmt.Errorf("parse video service doc: %w", err)
	}

	return data.Stitchedstream.Source, nil
}

func uriPrefixFromMasterURL(masterURL string) (string, error) {
	if s, _, found := strings.Cut(masterURL, "/master.m3u8?"); found {
		return s + "/", nil
	} else {
		return "", errors.New("invalid master URL: does not point to master.m3u8")
	}
}

type HLSFormat struct {
	AverageBandwidth uint
	FrameRate        float32
	Codecs           string
	Width            uint
	Height           uint
	Bandwidth        uint
	URI              string
}

type HLSMaster struct {
	AudioURI     string
	SubsURI      string
	VideoFormats []HLSFormat
}

// Formats are returned sorted from best to worst
func parseMasterM3U8(ctx context.Context, url string) (HLSMaster, error) {
	body, err := httputils.GetBodyWithContext(ctx, url)
	if err != nil {
		return HLSMaster{}, fmt.Errorf("get master HLS playlist: %w", err)
	}
	lines := strings.Split(string(body), "\n")

	var res HLSMaster
	var format HLSFormat
	for _, line := range lines {
		if mediaInfoStr, found := cutPrefix(line, "#EXT-X-MEDIA:"); found {
			var mediaInfo struct {
				Type       string
				GroupID    string
				Name       string
				URI        string
				AutoSelect string
			}

			err := getExtM3UInfo(mediaInfoStr, &mediaInfo)
			if err != nil {
				return HLSMaster{}, fmt.Errorf("parse EXTM3U: %w", err)
			}

			if mediaInfo.AutoSelect == "YES" {
				switch mediaInfo.Type {
				case "AUDIO":
					res.AudioURI = mediaInfo.URI
				case "SUBTITLES":
					res.SubsURI = mediaInfo.URI
				}
			}
		} else if streamInfoStr, found := cutPrefix(line, "#EXT-X-STREAM-INF:"); found {
			var streamInfo struct {
				AverageBandwidth int
				FrameRate        float32
				Codecs           string
				Resolution       string
				Bandwidth        int
			}

			err := getExtM3UInfo(streamInfoStr, &streamInfo)
			if err != nil {
				return HLSMaster{}, fmt.Errorf("parse EXTM3U: %w", err)
			}

			format.AverageBandwidth = uint(streamInfo.AverageBandwidth)
			format.FrameRate = streamInfo.FrameRate
			format.Codecs = streamInfo.Codecs
			{
				sp := strings.SplitN(streamInfo.Resolution, "x", 2)
				if len(sp) != 2 {
					return HLSMaster{}, errors.New("invalid resolution format in EXT-X-STREAM-INF")
				}
				w, err := strconv.ParseUint(sp[0], 10, 32)
				if err != nil {
					return HLSMaster{}, fmt.Errorf("parse resolution width: %w", err)
				}
				format.Width = uint(w)
				h, err := strconv.ParseUint(sp[1], 10, 32)
				if err != nil {
					return HLSMaster{}, fmt.Errorf("parse resolution height: %w", err)
				}
				format.Height = uint(h)
			}
			format.Bandwidth = uint(streamInfo.Bandwidth)
		} else if line != "" && !strings.HasPrefix(line, "#") {
			format.URI = line
			res.VideoFormats = append(res.VideoFormats, format)
			format = HLSFormat{}
		}
	}

	// Sort by bandwidth (best first)
	sort.Slice(res.VideoFormats, func(i, j int) bool {
		return res.VideoFormats[j].Bandwidth < res.VideoFormats[i].Bandwidth
	})

	return res, nil
}

type HLSStreamKey struct {
	Method string
	Key    []byte
}

type HLSStreamSegment struct {
	Duration float64
	URL      string
}

type HLSStream struct {
	Key      *HLSStreamKey
	Segments []HLSStreamSegment
}

func getHLSStream(ctx context.Context, uriPrefix, hlsURI string) (HLSStream, error) {
	body, err := httputils.GetBodyWithContext(ctx, uriPrefix+hlsURI)
	if err != nil {
		return HLSStream{}, fmt.Errorf("get stream HLS playlist: %w", err)
	}
	lines := strings.Split(string(body), "\n")

	var keyInfo struct {
		Method string
		URI    string
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

			// Each URI has <uriPrefix>/<the same starting sequence>/<filename>.
			// This starting sequence isn't included with the key, so we get it here.
			keyPfx, _, found := strings.Cut(hlsURI, "/")
			if !found {
				return HLSStream{}, fmt.Errorf("malformed HLS URI: missing '/': '%v'", hlsURI)
			}
			keyPfx += "/"

			// Download key
			key, err = httputils.GetBodyWithContext(ctx, uriPrefix+keyPfx+keyInfo.URI)
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

	res := HLSStream{
		Segments: segments,
	}

	if key != nil {
		res.Key = &HLSStreamKey{
			Method: keyInfo.Method,
			Key:    key,
		}
	}

	return res, nil
}

type EpisodeStream struct {
	Video HLSStream
	Audio HLSStream
	Subs  HLSStream
}

func GetEpisodeStream(ctx context.Context, e Episode, selectFormat func([]HLSFormat) (HLSFormat, error)) (EpisodeStream, error) {
	mediaMasterURL, err := getMediaMasterURL(ctx, e.URL)
	if err != nil {
		return EpisodeStream{}, fmt.Errorf("getMediaMasterURL: %w", err)
	}

	uriPrefix, err := uriPrefixFromMasterURL(mediaMasterURL)
	if err != nil {
		return EpisodeStream{}, err
	}

	hlsMaster, err := parseMasterM3U8(ctx, mediaMasterURL)
	if err != nil {
		return EpisodeStream{}, fmt.Errorf("parseMasterM3U8: %w", err)
	}

	videoFormat, err := selectFormat(hlsMaster.VideoFormats)
	if err != nil {
		return EpisodeStream{}, fmt.Errorf("selectFormat: %w", err)
	}

	videoStream, err := getHLSStream(ctx, uriPrefix, videoFormat.URI)
	if err != nil {
		return EpisodeStream{}, fmt.Errorf("getHLSStream: %w", err)
	}
	if videoStream.Key == nil || videoStream.Key.Method != "AES-128" {
		return EpisodeStream{}, fmt.Errorf("unable to decrypt '%v'; only AES-128 decryption is supported", videoStream.Key.Method)
	}

	audioStream, err := getHLSStream(ctx, uriPrefix, hlsMaster.AudioURI)
	if err != nil {
		return EpisodeStream{}, fmt.Errorf("getHLSStream: %w", err)
	}
	if audioStream.Key == nil || audioStream.Key.Method != "AES-128" {
		return EpisodeStream{}, fmt.Errorf("unable to decrypt '%v'; only AES-128 decryption is supported", videoStream.Key.Method)
	}

	subsStream, err := getHLSStream(ctx, uriPrefix, hlsMaster.SubsURI)
	if err != nil {
		return EpisodeStream{}, fmt.Errorf("getHLSStream: %w", err)
	}
	if subsStream.Key != nil {
		return EpisodeStream{}, fmt.Errorf("expected subs to be unencrypted, but found '%v' key", subsStream.Key.Method)
	}

	return EpisodeStream{
		Video: videoStream,
		Audio: audioStream,
		Subs:  subsStream,
	}, nil
}

// Download order: video, audio, subs. Does not interleave different media types.
func DownloadEpisodeStream(
	ctx context.Context,
	stream EpisodeStream,
	startSegment int,
	totalSegmentIdxCallback func(segmentIdx int),
	videoCallback func(data []byte, videoSegmentIdx int) error,
	audioCallback func(data []byte, audioSegmentIdx int) error,
	subsCallback func(data []byte, subsSegmentIdx int) error,
) error {
	segmentIndex := startSegment
	segmentOffset := 0

	totalSegmentIdxCallback(segmentIndex)

	for segmentIndex-segmentOffset < len(stream.Video.Segments) {
		relSegIdx := segmentIndex - segmentOffset
		seg := stream.Video.Segments[relSegIdx]
		data, err := downloadAndDecryptAES128Segment(ctx, seg.URL, stream.Video.Key.Key, relSegIdx)
		if err != nil {
			return fmt.Errorf("downloadAndDecryptAES128Segment (video): %w", err)
		}
		if err := videoCallback(data, relSegIdx); err != nil {
			return fmt.Errorf("videoCallback: %w", err)
		}
		segmentIndex++
		totalSegmentIdxCallback(segmentIndex)
	}
	segmentOffset += len(stream.Video.Segments)

	for segmentIndex-segmentOffset < len(stream.Audio.Segments) {
		relSegIdx := segmentIndex - segmentOffset
		seg := stream.Audio.Segments[relSegIdx]
		data, err := downloadAndDecryptAES128Segment(ctx, seg.URL, stream.Audio.Key.Key, relSegIdx)
		if err != nil {
			return fmt.Errorf("downloadAndDecryptAES128Segment (audio): %w", err)
		}
		if err := audioCallback(data, relSegIdx); err != nil {
			return fmt.Errorf("audioCallback: %w", err)
		}
		segmentIndex++
		totalSegmentIdxCallback(segmentIndex)
	}
	segmentOffset += len(stream.Audio.Segments)

	for segmentIndex-segmentOffset < len(stream.Subs.Segments) {
		relSegIdx := segmentIndex - segmentOffset
		seg := stream.Subs.Segments[relSegIdx]
		data, err := httputils.GetBodyWithContext(ctx, seg.URL)
		if err != nil {
			return fmt.Errorf("download subtitle segment: %w", err)
		}
		if err := subsCallback(data, relSegIdx); err != nil {
			return fmt.Errorf("subsCallback: %w", err)
		}
		segmentIndex++
		totalSegmentIdxCallback(segmentIndex)
	}

	return nil
}
