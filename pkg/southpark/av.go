package southpark

import (
	"errors"
	"fmt"
	"io"

	"github.com/yapingcat/gomedia/go-codec"
	"github.com/yapingcat/gomedia/go-mp4"
	"github.com/yapingcat/gomedia/go-mpeg2"
)

func ConvertTSAndAACToMP4(tsInput io.Reader, aacInput io.Reader, mp4Output io.WriteSeeker) error {
	var nextAACFrame func() ([]byte, uint64, uint64, bool)
	{
		const samplesPerFrame = 1024
		const sampleRate = 48000
		const millisecondsPerFrame = float64(samplesPerFrame) / float64(sampleRate) * 1000

		aacData, err := io.ReadAll(aacInput)
		if err != nil {
			return fmt.Errorf("read aac: %w", err)
		}

		var aacFrameBuf [][]byte
		codec.SplitAACFrame(aacData, func(frame []byte) {
			aacFrameBuf = append(aacFrameBuf, frame)
		})

		pts := float64(0) // in ms
		dts := float64(0) // in ms
		nextAACFrame = func() ([]byte, uint64, uint64, bool) {
			if len(aacFrameBuf) == 0 {
				return nil, uint64(pts), uint64(dts), false
			}

			defer func() {
				pts += millisecondsPerFrame
				dts += millisecondsPerFrame
			}()

			res := aacFrameBuf[0]
			aacFrameBuf = aacFrameBuf[1:]
			return res, uint64(pts), uint64(dts), true
		}
	}

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
	prevADTS := uint64(0)
	tsDemuxer := mpeg2.NewTSDemuxer()
	tsDemuxer.OnFrame = func(cid mpeg2.TS_STREAM_TYPE, vframe []byte, vpts uint64 /* in ms */, vdts uint64 /* in ms */) {
		if cid == mpeg2.TS_STREAM_H264 {
			if !hasVideo {
				vtid = muxer.AddVideoTrack(mp4.MP4_CODEC_H264)
				hasVideo = true
			}
			if err := muxer.Write(vtid, vframe, uint64(vpts), uint64(vdts)); err != nil {
				writeErr = err
			}

			for vdts > prevADTS {
				if !hasAudio {
					atid = muxer.AddAudioTrack(mp4.MP4_CODEC_AAC)
					hasAudio = true
				}
				if aframe, apts, adts, ok := nextAACFrame(); ok {
					if err := muxer.Write(atid, aframe, uint64(apts), uint64(adts)); err != nil {
						writeErr = err
					}
					prevADTS = adts
				} else {
					break
				}
			}
		}
	}

	if err := tsDemuxer.Input(tsInput); err != nil {
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
