package models_test

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/model"
)

func TestGameFromCSV(t *testing.T) {
	const input = `1262350,"SIGNALIS","Oct 27, 2022","20000 - 50000",0,0,19.99,0,0,"Searching for dreams inside of a nightmare Awaken from slumber and explore a surreal retrotech world as Elster, a technician Replika searching for her lost partner and her lost dreams. Discover terrifying secrets, challenging puzzles, and nightmarish creatures in a tense and melancholic experience of cosmic dread and classic psychological survival horror. REMEMBER OUR PROMISE FATAL EXCEPTION SIGNALIS is set in a dystopian future where humanity has colonized the solar system, and the totalitarian regime of Eusan maintains an iron grip through aggressive surveillance and propaganda. Humanoid androids known as Replikas live among the populace, acting as workers, civil servants, and Protektors alongside the citizens they are designed to resemble. The story of SIGNALIS begins as a Replika named Elster awakens from cryostasis in a wrecked vessel. Now stranded on a cold planet, she sets out on a journey into depths unknown. CLASSIC SURVIVAL HORROR GAMEPLAY Experience fear and apprehension as you encounter strange horrors, carefully manage scarce resources, and seek solutions to challenging riddles. COLD AND DISTANT PLACES Explore the dim corners of a derelict spaceship, delve into the mysterious fate of the inhabitants of a doomed facility, and seek what lies beneath. A DREAM ABOUT DREAMING Discover an atmospheric science-fiction tale of identity, memory, and the terror of the unknown and unknowable, inspired by classic cosmic horror and the works of Stanley Kubrick, Hideaki Anno, and David Lynch. A STRIKING VISION Wander a brutalist nightmare driven by fluid 3D character animations, dynamic lights and shadows, and complex transparency effects, complemented by cinematic sci-fi anime storytelling.","['English', 'German', 'Japanese', 'Korean', 'Russian', 'Simplified Chinese', 'Spanish - Latin America', 'French']","[]","","https://cdn.akamai.steamstatic.com/steam/apps/1262350/header.jpg?t=1666983782","http://rose-engine.org/signalis/","http://rose-engine.org","help@rose-engine.org",True,False,False,0,"",0,584,12,"",13,638,"This Game may contain content not appropriate for all ages, or may not be appropriate for viewing at work: Frequent Violence or Gore, General Mature Content.",0,0,0,0,"rose-engine","Humble Games,PLAYISM","Single-player,Full controller support","Action,Adventure,Indie","Survival Horror,Action,Sci-fi,Adventure,Female Protagonist,Violent,Atmospheric,Psychological Horror,Mystery,Cyberpunk,Dark,Lovecraftian,Dystopian,Singleplayer,Puzzle,Action-Adventure,Shooter,Third-Person Shooter,3D,Cinematic","https://cdn.akamai.steamstatic.com/steam/apps/1262350/ss_a2603694154878b8260c1dd498a06168cad012a4.1920x1080.jpg?t=1666983782,https://cdn.akamai.steamstatic.com/steam/apps/1262350/ss_9d602346170b19121e4baec94f5fab54cc43637c.1920x1080.jpg?t=1666983782,https://cdn.akamai.steamstatic.com/steam/apps/1262350/ss_dc87789ecfa2f83dd58ac778866fbf7fdc1ec99a.1920x1080.jpg?t=1666983782,https://cdn.akamai.steamstatic.com/steam/apps/1262350/ss_bb676746c44db01c2bd7ca78616032984f958106.1920x1080.jpg?t=1666983782,https://cdn.akamai.steamstatic.com/steam/apps/1262350/ss_4ec0c13288054a99bc3987680a914ccda834e7f6.1920x1080.jpg?t=1666983782,https://cdn.akamai.steamstatic.com/steam/apps/1262350/ss_f7eb1a1944c4c9d3142bedb800ad3a43a6f82a60.1920x1080.jpg?t=1666983782,https://cdn.akamai.steamstatic.com/steam/apps/1262350/ss_947db45ff0fb6da0e2d9031f0f71a32b740ed612.1920x1080.jpg?t=1666983782,https://cdn.akamai.steamstatic.com/steam/apps/1262350/ss_dcf8d44a07289d9b8ca9d1e9b9ae2a002dd76f6e.1920x1080.jpg?t=1666983782,https://cdn.akamai.steamstatic.com/steam/apps/1262350/ss_c7fcc30c5e2cd2ddf44bc01b2d43964abba72076.1920x1080.jpg?t=1666983782,https://cdn.akamai.steamstatic.com/steam/apps/1262350/ss_5338061a9fa31755789e0a7698fadc2d4ab29b94.1920x1080.jpg?t=1666983782,https://cdn.akamai.steamstatic.com/steam/apps/1262350/ss_af2cb31dc253d739c3ce2f6930be5d339e199be2.1920x1080.jpg?t=1666983782","http://cdn.akamai.steamstatic.com/steam/apps/256913244/movie_max.mp4?t=1666887901,http://cdn.akamai.steamstatic.com/steam/apps/256910985/movie_max.mp4?t=1665773311,http://cdn.akamai.steamstatic.com/steam/apps/256891139/movie_max.mp4?t=1664497433,http://cdn.akamai.steamstatic.com/steam/apps/256878180/movie_max.mp4?t=1664497256"
`

	r := csv.NewReader(strings.NewReader(input))
	r.LazyQuotes = true
	csvLine, err := r.Read()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	got, err := models.GameFromCSVLine(csvLine)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	want := &models.Game{
		AppID:       "1262350",
		Name:        "SIGNALIS",
		Genres:      "Action,Adventure,Indie",
		ReleaseYear: 2010,
		AvgPlayTime: 0.0,
		SupportedOS: models.WindowsMask | (^models.LinuxMask) | (^models.MacMask),
	}
	if *got != *want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestReviewFromCSV(t *testing.T) {
	const input = `8870,BioShock Infinite,"By playing this game and finally understanding it, my mind lost its virginity. 10/10 would totally lose mind virginity again",1,0`
	r := csv.NewReader(strings.NewReader(input))
	r.LazyQuotes = true
	csvLine, err := r.Read()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	got, err := models.ReviewFromCSVLine(csvLine)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	want := &models.Review{
		AppID: "8870",
		Text:  "By playing this game and finally understanding it, my mind lost its virginity. 10/10 would totally lose mind virginity again",
		Score: models.Positive,
	}
	
	if *got != *want {
		t.Fatalf("got %v, want %v", got, want)
	}
}
