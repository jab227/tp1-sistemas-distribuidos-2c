package main

import (
	"fmt"
	"github.com/pemistahl/lingua-go"
)

func main() {
	languages := []lingua.Language{
		lingua.English,
		lingua.Spanish,
	}
	// Create a language detector for English
	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages(languages...).
		Build()

	text := "This is a sample English text."

	// Detect the language of the text
	detectedLanguage, exists := detector.DetectLanguageOf(text)

	if exists && detectedLanguage == lingua.English {
		fmt.Println("The text is in English.")
	} else {
		fmt.Println("The text is not in English.")
	}
}
