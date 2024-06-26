package southpark

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/yapingcat/gomedia/go-codec"
	"github.com/yapingcat/gomedia/go-mp4"
	"github.com/yapingcat/gomedia/go-mpeg2"
)

type SegmentFile struct {
	Filename string
	Duration float64
}

func ConvertTSAndAACToMP4(tsInput []SegmentFile, aacInput []SegmentFile, mp4Output io.WriteSeeker, onProgress func(progress float64)) error {
	//var _vdts uint64
	//var _vpts uint64

	muxer, err := mp4.CreateMp4Muxer(mp4Output)
	if err != nil {
		return fmt.Errorf("create mp4 muxer: %w", err)
	}

	var onFrameErr error

	const aacSamplesPerFrame = 1024
	const aacSampleRate = 48000
	const aacMillisecondsPerFrame = float64(aacSamplesPerFrame) / float64(aacSampleRate) * 1000
	apts := float64(0)
	adts := float64(0)
	var aacFrameBuf [][]byte

	// https://github.com/yapingcat/gomedia/blob/main/example/example_convert_ts_to_mp4.go
	hasAudio := false
	hasVideo := false
	var atid uint32 = 0
	var vtid uint32 = 0
	prevADTS := float64(0)
	tsDemuxer := mpeg2.NewTSDemuxer()
	tsDemuxer.OnFrame = func(cid mpeg2.TS_STREAM_TYPE, vframe []byte, vpts uint64 /* in ms */, vdts uint64 /* in ms */) {
		if cid == mpeg2.TS_STREAM_H264 {
			//for uint64(float64(vdts) * 1.000642) > prevADTS {
			for float64(vdts) > prevADTS {
				if !hasAudio {
					atid = muxer.AddAudioTrack(mp4.MP4_CODEC_AAC)
					hasAudio = true
				}
				if len(aacFrameBuf) > 0 {
					if err := muxer.Write(atid, aacFrameBuf[0], uint64(apts), uint64(adts)); err != nil {
						onFrameErr = err
					}
					prevADTS = adts
				} else {
					break
				}
				aacFrameBuf = aacFrameBuf[1:]
				apts += aacMillisecondsPerFrame
				adts += aacMillisecondsPerFrame
			}

			if !hasVideo {
				vtid = muxer.AddVideoTrack(mp4.MP4_CODEC_H264)
				hasVideo = true
			}
			if err := muxer.Write(vtid, vframe, uint64(vpts), uint64(vdts)); err != nil {
				onFrameErr = err
			}
			//_vpts = vpts
			//_vdts = vdts
		}
	}

	if len(aacInput) != len(tsInput) {
		return fmt.Errorf("number of AAC segments (%v) and TS segments (%v) doesn't match", len(aacInput), len(tsInput))
	}

	var aHLSTime float64
	for i := range tsInput {
		adts = aHLSTime * 1000
		apts = aHLSTime * 1000
		{
			data, err := os.ReadFile(aacInput[i].Filename)
			if err != nil {
				return err
			}
			codec.SplitAACFrame(data, func(frame []byte) {
				aacFrameBuf = append(aacFrameBuf, frame)
			})
		}
		{
			data, err := os.ReadFile(tsInput[i].Filename)
			if err != nil {
				return err
			}
			if err := tsDemuxer.Input(bytes.NewReader(data)); err != nil {
				if errors.Is(err, io.ErrUnexpectedEOF) {
					// File is incomplete, ignore
				} else {
					return fmt.Errorf("MPEG-TS demuxer: %w", err)
				}
			}
		}
		onProgress(float64(i) / float64(len(tsInput)))
		aHLSTime += aacInput[i].Duration
	}

	muxer.WriteTrailer()

	//fmt.Printf("**** %f %v %v %f %f\n", adts, _vpts, _vdts, float64(_vdts) - adts, float64(_vdts) / adts)

	if onFrameErr != nil {
		return fmt.Errorf("mp4 muxer: %w", onFrameErr)
	}

	return nil
}
