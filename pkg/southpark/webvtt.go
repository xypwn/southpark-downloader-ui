package southpark

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func mergeVTTSubtitles(parts [][]byte, durations []float64) ([]byte, error) {
	if len(parts) != len(durations) {
		return nil, fmt.Errorf("number of parts (%v) doesn't match number of durations (%v)", len(parts), len(durations))
	}

	var tOffset float64

	var res bytes.Buffer
	res.WriteString("WEBVTT\n\n")

	for i, part := range parts {
		if err := parseVTTSubtitles(
			part,
			func(comment string) {
				res.WriteString(fmt.Sprintf("NOTE Part %v %v\n\n", i+1, comment))
			},
			func(id string, tStart float64, tEnd float64, settings []string, transcript string) {
				timestampString := func(t float64) string {
					var res strings.Builder
					hrs := int64(t / (60 * 60))
					res.WriteString(fmt.Sprintf("%02d", hrs))
					t -= float64(hrs) * 60 * 60
					res.WriteRune(':')

					mins := int64(t / 60)
					res.WriteString(fmt.Sprintf("%02d", mins))
					t -= float64(mins) * 60
					res.WriteRune(':')

					secs := t
					res.WriteString(fmt.Sprintf("%06.3f", secs))
					t -= float64(mins)
					return res.String()
				}

				if id != "" {
					res.WriteString(fmt.Sprintf("%v:%v\n", i+1, id))
				}

				res.WriteString(fmt.Sprintf(
					"%v --> %v\n%v\n\n",
					timestampString(tStart+tOffset),
					timestampString(tEnd+tOffset),
					transcript,
				))
			},
			func(comment string) {
				res.WriteString(fmt.Sprintf("NOTE %v\n\n", comment))
			},
		); err != nil {
			return nil, fmt.Errorf("parseVTTSubtitles: %w", err)
		}

		// Ensure the next part's timestamps are offset properly
		tOffset += durations[i]
	}
	return res.Bytes(), nil
}

// Super simple and dumb WebVTT parser, not for general use
func parseVTTSubtitles(
	input []byte,
	onVTTHeader func(comment string),
	onCue func(id string, tStart float64, tEnd float64, settings []string, transcript string),
	onComment func(comment string),
) error {
	type Mode int
	const (
		ModeStart Mode = iota
		ModeWebVTTHeader
		ModeComment
		ModeCue
		ModeCueTranscript
	)
	mode := ModeStart

	var cueID string
	var cueStart, cueEnd float64
	var cueSettings []string
	text := strings.Builder{}

	lines := strings.Split(string(input), "\n")
	for _, line := range lines {
		switch {
		case mode == ModeStart:
			if s, ok := cutPrefix(line, "WEBVTT"); ok {
				text.WriteString(s)
			} else {
				return fmt.Errorf("expected 'WEBVTT', but got '%v'", line)
			}
			mode = ModeWebVTTHeader
		case line == "":
			switch mode {
			case ModeWebVTTHeader:
				onVTTHeader(text.String())
			case ModeComment:
				onComment(text.String())
			case ModeCueTranscript:
				onCue(cueID, cueStart, cueEnd, cueSettings, text.String())
				cueID = ""
				cueStart = 0
				cueEnd = 0
				cueSettings = nil
			}
			text.Reset()
			mode = ModeCue
		case strings.HasPrefix(line, "NOTE"):
			text.WriteString(strings.TrimPrefix(strings.TrimPrefix(line, "NOTE"), " "))
			mode = ModeComment
		case mode == ModeWebVTTHeader || mode == ModeComment || mode == ModeCueTranscript:
			if text.Len() > 0 {
				text.WriteRune('\n')
			}
			text.WriteString(line)
		case mode == ModeCue:
			if strings.Contains(line, " --> ") {
				sp := strings.Split(line, " ")
				if len(sp) < 3 || sp[1] != "-->" {
					return fmt.Errorf("expected '<tStart> --> <tEnd> [settings...]', but got '%v'", line)
				}

				parseTimestamp := func(s string) (float64, error) {
					var res float64
					var multiplier float64 = 1
					sp := strings.Split(s, ":")
					for i := len(sp) - 1; i >= 0; i-- {
						x, err := strconv.ParseFloat(sp[i], 64)
						if err != nil {
							return 0, err
						}
						res += x * multiplier
						multiplier *= 60
					}
					return res, nil
				}

				start, err := parseTimestamp(sp[0])
				if err != nil {
					return fmt.Errorf("parsing start timestamp: %w", err)
				}

				end, err := parseTimestamp(sp[2])
				if err != nil {
					return fmt.Errorf("parsing end timestamp: %w", err)
				}

				cueStart = start
				cueEnd = end
				cueSettings = sp[3:]

				mode = ModeCueTranscript
			} else {
				cueID = line
			}
		}
	}

	return nil
}
