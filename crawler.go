package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

// Link struct represents a page and all links found on the page
type Link struct {
	URL   *url.URL
	links []*Link
}

func main() {
	// Setup logger
	f, err := os.OpenFile("run.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	// Define and save command line args
	baseURLStr := flag.String("b", "monzo.com", "Starting URL to crawl from")
	var searchDepth int
	flag.IntVar(&searchDepth, "d", 5, "Number of levels you want to traverse (depth)")
	flag.Parse()

	// Start the crawl
	log.Print(fmt.Sprintf("------- STARTING NEW CRAWL FOR: %s -------", *baseURLStr))

	baseURL, err := parseURL(*baseURLStr)
	if err != nil {
		logger("e", "Base url could not be parsed")
		panic(err)
	}

	// create link and waitgroup
	baseLink := &Link{URL: baseURL}
	var globalWait sync.WaitGroup

	// start crawl from base link
	globalWait.Add(1)
	go crawl(baseLink, searchDepth, baseURL, &globalWait)
	globalWait.Wait()

	// all routines returned so we can now print the textual sitemap
	outputFile, err := os.Create(fmt.Sprintf("%s.txt", strings.Split(baseURL.Hostname(), ".")[0]))
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	printSitemap(baseLink, 0, outputFile)

}

/*
Crawls one page:
1) check if depth reached, and return if it has
2) gets all links on a page
3) for each link
	. recursively set a new go routine running to crawl
*/
func crawl(link *Link, depth int, initialBaseURL *url.URL, globalWait *sync.WaitGroup) {
	logger("i", fmt.Sprintf("Starting crawl for: %s  ||  depth of %d", link.URL.String(), depth))
	// handle wait group at start
	defer globalWait.Done()

	// check depth
	if depth == 0 {
		return
	}

	// get all links on the given page
	newLinks, err := getLinksFromURL(link.URL, initialBaseURL)
	if err != nil {
		logger("e", fmt.Sprintf("Error getting links for: %s \n", link.URL.String()))
		return
	}

	for _, newL := range newLinks {
		// add to current links Links slice
		link.links = append(link.links, newL)
		// only proceed if next depth > 0
		if depth-1 > 0 {
			// Crawl each link found on this page
			globalWait.Add(1)
			go crawl(newL, depth-1, initialBaseURL, globalWait)
		}
	}
	return
}

func getLinksFromURL(link *url.URL, baseURL *url.URL) ([]*Link, error) {
	// get response
	resp, err := http.Get(link.String())
	if err != nil {
		logger("e", fmt.Sprintf("Error getting response from: %s", link.String()))
		return nil, err
	}

	// parse for <a> tags
	var newLinks []*Link
	z := html.NewTokenizer(resp.Body)

	// used to avoid duplicate links being returned
	seen := make(map[string]bool)

	for {
		token := z.Next()

		switch {
		case token == html.ErrorToken:
			// End of page, return
			return newLinks, nil
		case token == html.StartTagToken:
			// check if anchor tag found
			tag := z.Token()
			// worth noting  here thatthis does not guarantee 100% of links, could be stuff in javascript somewhere
			isAnchor := tag.Data == "a" //|| tag.Data == "link"
			if isAnchor {
				// get href attribute
				for _, a := range tag.Attr {
					if a.Key == "href" {

						// link found, lets check if it belongs to current subdomain
						l, err := parseURL(a.Val)
						if err != nil {
							logger("e", fmt.Sprintf("Error parsing link: %s", a.Val))
							return nil, err
						}

						if l.Hostname() == baseURL.Hostname() || !l.IsAbs() {
							// if link is a path, append the domain to it
							if !l.IsAbs() {
								l = baseURL.ResolveReference(l)
							}

							// check if link has been seen on this page allready
							if _, ok := seen[l.String()]; !ok {
								//link has not been seen before and is of the right domain
								lLink := &Link{URL: l}
								newLinks = append(newLinks, lLink)
								// add to seen map
								seen[l.String()] = true
								break
							}
						} else {
							logger("i", fmt.Sprintf("Discarding url: %s", l.String()))
						}
					}
				}
			}
		}
	}
}

// Takes a url string and retruns a url.URL type if valid
func parseURL(URLString string) (*url.URL, error) {
	resultURL, err := url.Parse(URLString)
	if err != nil {
		logger("e", fmt.Sprintf("Could not parse URL: %s", URLString))
	}
	return resultURL, err
}

// Prints all links in the baseLink with the given indent level
// To be used recursively
func printSitemap(baseLink *Link, indent int, outputFile *os.File) error {
	indentString := "        "

	// print first link
	if indent == 0 {
		_, err := io.WriteString(outputFile, fmt.Sprintf("%s\n", baseLink.URL.String()))
		if err != nil {
			logger("e", fmt.Sprintf("Failed to write to file for: %s", baseLink.URL.String()))
		}
	}

	for _, l := range baseLink.links {
		_, err := io.WriteString(outputFile, fmt.Sprintf("%s - %s\n", strings.Repeat(indentString, indent), l.URL.String()))
		if err != nil {
			logger("e", fmt.Sprintf("Failed to write to file for: %s", l.URL.String()))
		}
		// print children links
		printSitemap(l, indent+1, outputFile)
	}
	return nil
}

// Allows for levels of severity in our logger
func logger(severity string, message string) {
	switch severity {
	case "e":
		log.Printf("[ERROR] %s", message)
	case "i":
		log.Printf("[INFO] %s", message)
	}
}
