package utils_test

import (
	"github.com/jab227/tp1-sistemas-distribuidos-2c/internal/utils"
	"strings"
	"testing"
)

func TestReadLines(t *testing.T) {
	const input = `227300,Euro Truck Simulator 2,"When you need to take a good, solid break from all the outside noise of life (and maybe even other games [See my CS: GO review]), this is your solution. Drive around Great Britain and continental Europe and see it in its innocence. The objective is to run a successful trucking company, so be patient when starting, as the workload will get lighter as time goes on.",1,0
227300,Euro Truck Simulator 2,The Turkish guys from here are like the Russians from CS: GO...,1,0
227300,Euro Truck Simulator 2,"Having played Euro Truck Simulator 2 with my headset firmly lodged on my head, listening to my regional radio stations from the in-game radio, I can safely say that playing this game is one of the best escapes I've experienced in gaming.  I enjoy playing Skyrim, Grand Theft Auto, Borderlands, CS: GO, etc., but none of these games quite give the same relaxation as Euro Truck Simulator 2 does. For a large part of the time, it's almost an excercise in quiet, calm reflection over the world's wonders, as you drift into auto-pilot and consume the commercial hallmarks of real-life radio advertisements.  You may think the premise is boring, but even if you're doubtful about the *wonderful* cruising from A to B (and then C), you'll even get to practice hiring other guys to drive for you. Which, to be honest, you don't want them to do too much, because *driving it yourself is so enjoyable*!",1,0
`
	inputSplit := strings.Split(input, "\n")

	r := strings.NewReader(input)
	lr := utils.NewLinesReader(r, 2)

	lines, more, err := lr.Next()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(lines) != 2 {
		t.Errorf("unexpected len: got %d, want %v", len(lines), 2)
	}

	if lines[0] != inputSplit[0] {
		t.Errorf("unexpected line: got %q, want %q", lines[0], inputSplit[0])
	}

	if lines[1] != inputSplit[1] {
		t.Errorf("unexpected line: got %q, want %q", lines[1], inputSplit[1])
	}

	if !more {
		t.Error("expected more elements")
	}
	
	lines, more, err = lr.Next()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if len(lines) != 1 {
		t.Errorf("unexpected len: got %d, want %d", len(lines), 1)
	}

	if lines[0] != inputSplit[2] {
		t.Errorf("unexpected line: got %q, want %q", lines[0], inputSplit[2])
	}

	if more {
		t.Error("expected no more elements")
	}
	
	lines, more, err = lr.Next()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if more {
		t.Error("expected no more elemnts")
	}
}
