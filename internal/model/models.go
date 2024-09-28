package models

// Position in the csv file counting from zero

import (
	"fmt"
	"strconv"
	"time"
)

const releaseDateFmt = "Jan 02, 2006"

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

func GameFromCSVLine(csvLine []string) (*Game, error) {
	releaseDate, err := time.Parse(releaseDateFmt, csvLine[ReleaseDateCSVPosition])
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

type ReviewScore int8

const (
	Positive ReviewScore = 1
	Negative ReviewScore = -1
)

func reviewScoreFromString(s string) (ReviewScore, error){
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
		Text:  csvLine[ReviewTextCSVPosition],
		Score: reviewScore,
	}, nil
}
