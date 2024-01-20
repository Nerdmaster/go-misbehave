package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

var commonWords []string

func main() {
	var duration = flag.Int("duration", 10, "Duration in seconds")
	var byteLimit = flag.Int("bytes", 1024, "Byte limit")
	var bind = flag.String("bind", ":8080", "Bind address")
	flag.Parse()

	loadWords()

	var connections atomic.Int64
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		connections.Add(1)
		var connum = connections.Load()
		var ip = strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0])
		var ua = r.Header.Get("User-Agent")
		log.Printf("Connection %d started from %q: %s %s (UA: %q)", connum, ip, r.Method, r.RequestURI, ua)
		generateHTML(w, time.Second*time.Duration(*duration), *byteLimit)
		log.Printf("Connection %d complete", connum)
	})

	log.Printf("Server listening on %q", *bind)
	log.Printf("Will listen and respond with random HTML for duration %d seconds or %d bytes, whichever comes first.", *duration, *byteLimit)
	var err = http.ListenAndServe(*bind, nil)
	if err != nil {
		log.Fatalf("Unable to start server on %q: %s", *bind, err)
	}
}

func loadWords() {
	var file, err = os.Open("common.txt")
	if err != nil {
		log.Fatalf("Unable to open common.txt wordlist: %s", err)
	}
	defer file.Close()

	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		commonWords = append(commonWords, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Unable to read common.txt: %s", err)
	}
}

func generateHTML(w http.ResponseWriter, duration time.Duration, byteLimit int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<!DOCTYPE html>\n<html>\n<head>\n  <title>%s</title>\n</head>\n<body>\n", randContent())

	var startTime = time.Now()
	var totalBytes = 0
	var openTags []string

	// HTML generation loop
	for totalBytes < byteLimit && time.Since(startTime) < duration {
		var numTags = len(openTags)

		// Open a tag unless we have a bunch, with odds decreasing the more tags we already have
		var odds = 15 - numTags
		if rand.Intn(10) < odds {
			var output, newtag = randomTag()
			var indent = strings.Repeat("  ", len(openTags)) // Two spaces per open tag
			var n, err = fmt.Fprint(w, indent+output)
			if err != nil {
				log.Printf("Error writing to output stream: %s", err)
				return
			}
			totalBytes += n
			if newtag != "" {
				openTags = append(openTags, newtag)
			}
		}

		// Close a tag unless we have very few, with odds increasing the more tags we already have
		odds = numTags - 2
		if rand.Intn(10) < odds {
			var lastTag = openTags[numTags-1]
			openTags = openTags[:numTags-1]
			var indent = strings.Repeat("  ", len(openTags))
			var n, err = fmt.Fprintf(w, "%s</%s>\n", indent, lastTag)
			if err != nil {
				log.Printf("Error writing to output stream: %s", err)
				return
			}
			totalBytes += n
		}

		time.Sleep(time.Millisecond * 1)
	}

	log.Printf("- Duration: %s", time.Since(startTime))
	log.Printf("- Bytes sent: %d", totalBytes)
	for i := len(openTags) - 1; i >= 0; i-- {
		var indent = strings.Repeat("  ", i)
		fmt.Fprintf(w, "%s</%s>\n", indent, openTags[i])
	}
	fmt.Fprint(w, "</body>\n</html>\n")
}

var contentTags = []string{"p", "div", "span", "a", "li", "h1", "h2", "h3"}
var elementTags = []string{"div", "ul"}

func randomTag() (output string, newtag string) {
	switch rand.Intn(10) {
	case 0, 1, 2, 3, 4, 5:
		var tag = contentTags[rand.Intn(len(contentTags))]
		var content = randContent()
		return fmt.Sprintf("<%s>%s</%s>\n", tag, content, tag), ""
	case 6, 7, 8:
		var tag = elementTags[rand.Intn(len(elementTags))]
		return fmt.Sprintf("<%s>", tag), tag
	case 9:
		var attr = randAttr()
		var val = randVal()
		var word1, word2, word3, word4 = randWord(), randWord(), randWord(), randWord()
		return fmt.Sprintf(`<a href="/misbehave/%s/%s/%s/%s" %s=%q>%s</a>`, word1, word2, word3, word4, attr, val, randContent()), ""
	default:
		return "", ""
	}
}

// Helper function to generate a random word-like string
func randWord() string {
	return commonWords[rand.Intn(len(commonWords))]
}

// randContent creates a string of random word-like text
func randContent() string {
	numWords := rand.Intn(20) + 10 // 11 to 30 words
	var content strings.Builder
	for i := 0; i < numWords; i++ {
		content.WriteString(randWord() + " ")
	}
	return strings.TrimSpace(content.String())
}

// randAttr returns a single random attribute name
func randAttr() string {
	attributes := []string{"class", "id", "data-dog", "src", "alt", "title", "style"}
	return attributes[rand.Intn(len(attributes))]
}

// randVal generates a short random fake word
func randVal() string {
	return randWord()
}
