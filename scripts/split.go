package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sync"
)

type Config struct {
	Filename string
	Prefix   string
	Count    int
	IsRandom bool
}

func GetConfig() Config {
	config := Config{}
	flag.StringVar(&config.Filename, "filename", "", "filename of the file to split")
	flag.StringVar(&config.Prefix, "outPrefix", "out", "prefix of the out files")
	flag.IntVar(&config.Count, "count", 3, "number of files in which to split the original")
	flag.BoolVar(&config.IsRandom, "rand", false, "if true the contents of each of the files will be random")
	flag.Parse()
	return config
}

func main() {
	config := GetConfig()
	var r io.Reader
	if config.Filename != "" {
		f, err := os.Open(config.Filename)
		defer f.Close()
		if err != nil {
			log.Fatalf("error while opening input file: %e", err)
		}
		r = f
	} else {
		r = os.Stdin
	}

	channels := make([]chan string, 0, config.Count)
	var wg sync.WaitGroup
	for i := 0; i < config.Count; i++ {
		outfile := fmt.Sprintf("%s_%d.csv", config.Prefix, i+1)
		f, err := os.Create(outfile)
		defer f.Close()
		if err != nil {
			log.Fatalf("error while opening out file: %e", err)
		}
		w := bufio.NewWriter(f)
		ch := make(chan string, 1)
		channels = append(channels, ch)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for l := range ch {
				if l == "" {
					break
				}
				w.WriteString(l)
				w.WriteByte('\n')
			}
			w.Flush()
		}()
	}

	scanner := bufio.NewScanner(r)
	const capacity = 128 * 1024
	buffer := make([]byte, capacity)
	scanner.Buffer(buffer, 1024*1024)
	scanner.Scan()
	firstLine := scanner.Text()
	for _, ch := range channels {
		ch <- firstLine
	}

	i := 0

	var nextChFunc func(int) int

	if config.IsRandom {
		nextChFunc = func(i int) int {
			return int(rand.Int63n(int64(config.Count)))
		}
	} else {

		nextChFunc = func(i int) int {
			return (i + 1) % config.Count
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		ch := channels[i]
		ch <- line
		i = nextChFunc(i)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %s\n", err)
	}

	for _, ch := range channels {
		close(ch)
	}
	wg.Wait()
}
