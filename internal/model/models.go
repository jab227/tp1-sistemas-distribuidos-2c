package models

// Position in the csv file counting from zero

import "time"

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

type Game struct {
	AppID       string
	Name        string
	Genres      string
	ReleaseYear uint32
	AvgPlayTime float32
	SupportedOS OS
}

func GameFromCSVLine(csvLine []string) (*Game, error) {
	return nil, nil
}

type ReviewScore int8

const (
	Positive ReviewScore = 1
	Negative ReviewScore = -1
)

type Review struct {
	AppID string
	Text  string
	Score ReviewScore
}

func ReviewFromCSVLine(csvLine []string) (*Review, error) {
	return nil, nil
}
