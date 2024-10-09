package models

// Position in the csv file counting from zero

import (
	"fmt"
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/protocol"
	"strconv"
	"time"
)

const releaseDateFmtWithDay = "Jan 2, 2006"
const releaseDateFmtOnlyMonthYear = "Jan 2006"

const (
	AppIDCSVPosition              = 0
	NameCSVPosition               = 1
	ReleaseDateCSVPosition        = 2
	OSWindowsCSVPosition          = 17
	OSMacCSVPosition              = 18
	OSLinuxCSVPosition            = 19
	AvgPlaytimeForeverCSVPosition = 29
	GenresCSVPosition             = 36
	ReviewTextCSVPosition         = 2
	ReviewScoreCSVPosition        = 3
)

type OS byte

const (
	WindowsMask OS = (1 << 0) // 0b001
	MacMask     OS = (1 << 1) // 0b010
	LinuxMask   OS = (1 << 2) // 0b100
)

func (o OS) IsWindowsSupported() bool {
	return (o & WindowsMask) == WindowsMask
}

func (o OS) IsMacSupported() bool {
	return (o & MacMask) == MacMask
}

func (o OS) IsLinuxSupported() bool {
	return (o & LinuxMask) == LinuxMask
}

type Game struct {
	AppID       string
	Name        string
	Genres      string
	ReleaseYear uint32
	AvgPlayTime float32
	SupportedOS OS
}

type playableIn struct {
	windows bool
	mac     bool
	linux   bool
}

func getSupportedOSs(p playableIn) OS {
	var os OS
	if p.windows {
		os |= WindowsMask
	}
	if p.mac {
		os |= MacMask
	}
	if p.linux {
		os |= LinuxMask
	}
	return os
}

func ParseDate(data string) (time.Time, error) {
	listOfFormats := []string{releaseDateFmtWithDay, releaseDateFmtOnlyMonthYear}

	for _, format := range listOfFormats {
		releaseDate, err := time.Parse(format, data)
		if err != nil {
			continue
		} else {
			return releaseDate, nil
		}
	}

	return time.Time{}, fmt.Errorf("error with release date format %s", data)
}

func GameFromCSVLine(csvLine []string) (*Game, error) {
	releaseDate, err := ParseDate(csvLine[ReleaseDateCSVPosition])
	if err != nil {
		return nil, fmt.Errorf("couldn't parse release date: %w", err)
	}
	avgPlaytime, err := strconv.ParseFloat(csvLine[AvgPlaytimeForeverCSVPosition], 32)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse average playtime forever: %w", err)
	}

	osWindows, err := strconv.ParseBool(csvLine[OSWindowsCSVPosition])
	if err != nil {
		return nil, fmt.Errorf("couldn't parse windows flag: %w", err)
	}

	osMac, err := strconv.ParseBool(csvLine[OSMacCSVPosition])
	if err != nil {
		return nil, fmt.Errorf("couldn't parse mac flag: %w", err)
	}

	osLinux, err := strconv.ParseBool(csvLine[OSLinuxCSVPosition])
	if err != nil {
		return nil, fmt.Errorf("couldn't parse linux flag: %w", err)
	}

	supportedOS := getSupportedOSs(playableIn{
		windows: osWindows,
		mac:     osMac,
		linux:   osLinux,
	})

	return &Game{
		AppID:       csvLine[AppIDCSVPosition],
		Name:        csvLine[NameCSVPosition],
		Genres:      csvLine[GenresCSVPosition],
		ReleaseYear: uint32(releaseDate.Year()),
		AvgPlayTime: float32(avgPlaytime),
		SupportedOS: supportedOS,
	}, nil
}

func (g *Game) BuildPayload(builder *protocol.PayloadBuffer) {
	builder.BeginPayloadElement()

	builder.WriteBytes([]byte(g.AppID))
	builder.WriteBytes([]byte(g.Name))
	builder.WriteBytes([]byte(g.Genres))
	builder.WriteUint32(g.ReleaseYear)
	builder.WriteFloat32(g.AvgPlayTime)
	builder.WriteByte(byte(g.SupportedOS))

	builder.EndPayloadElement()
}

type ReviewScore int8

const (
	Positive ReviewScore = 1
	Negative ReviewScore = -1
)

func reviewScoreFromString(s string) (ReviewScore, error) {
	if s == "1" {
		return Positive, nil
	}

	if s == "-1" {
		return Negative, nil
	}
	return 0, fmt.Errorf("unknown score: %q", s)
}

type Review struct {
	AppID string
	Name  string
	Text  string
	Score ReviewScore
}

func ReviewFromCSVLine(csvLine []string) (*Review, error) {
	reviewScore, err := reviewScoreFromString(csvLine[ReviewScoreCSVPosition])
	if err != nil {
		return nil, fmt.Errorf("couldn't parse review score: %w", err)
	}
	return &Review{
		AppID: csvLine[AppIDCSVPosition],
		Name:  csvLine[NameCSVPosition],
		Text:  csvLine[ReviewTextCSVPosition],
		Score: reviewScore,
	}, nil
}

func (r *Review) BuildPayload(builder *protocol.PayloadBuffer) {
	builder.BeginPayloadElement()

	builder.WriteBytes([]byte(r.AppID))
	builder.WriteBytes([]byte(r.Name))
	builder.WriteBytes([]byte(r.Text))
	builder.WriteByte(byte(r.Score))

	builder.EndPayloadElement()
}

func ReadGame(element *protocol.Element) Game {
	game := Game{
		AppID:       string(element.ReadBytes()),
		Name:        string(element.ReadBytes()),
		Genres:      string(element.ReadBytes()),
		ReleaseYear: element.ReadUint32(),
		AvgPlayTime: element.ReadFloat32(),
		SupportedOS: OS(element.ReadByte()),
	}
	return game
}

func ReadReview(element *protocol.Element) Review {
	return Review{
		AppID: string(element.ReadBytes()),
		Name:  string(element.ReadBytes()),
		Text:  string(element.ReadBytes()),
		Score: ReviewScore(element.ReadByte()),
	}
}

func (g Game) GetID() string {
	return g.AppID
}

func (r Review) GetID() string {
	return r.AppID
}
