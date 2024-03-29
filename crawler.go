// Main package containing scraper logic and cli
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

// Page struct represents a page and all links found on the page
type Page struct {
	URL   *url.URL
	links []*Page
}

func main() {
	// setup logger
	createDirIfNotExist("./output")
	createDirIfNotExist("output/logs")
	logFileName := generateDateFileName("output/logs/log_")

	f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			return
		}
	}(f)
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	// setup cli
	baseURLStr := flag.String("b", "https://example.com", "Starting URL to crawl from")
	var searchDepth int
	flag.IntVar(&searchDepth, "d", 5, "Number of levels you want to traverse (depth)")
	flag.Parse()

	// start crawling
	log.Printf("------- STARTING NEW CRAWL FOR: %s -------", *baseURLStr)
	baseURL, err := parseURL(*baseURLStr)
	if err != nil {
		logger("e", "Base url could not be parsed")
		panic(err)
	}

	// start crawl from base page
	basePage := &Page{URL: baseURL}
	var globalWait sync.WaitGroup
	globalWait.Add(1)
	go crawl(basePage, searchDepth, baseURL, &globalWait)
	globalWait.Wait()

	// all routines returned, so we can now print the textual sitemap
	// create output file
	createDirIfNotExist("./output")
	var filePrefix string
	switch strings.Split(baseURL.Hostname(), ".")[0] {
	case "www":
		filePrefix = strings.Split(baseURL.Hostname(), ".")[1]
	default:
		filePrefix = strings.Split(baseURL.Hostname(), ".")[0]
	}
	outputFile, err := os.Create(generateDateFileName(fmt.Sprintf("output/%s_", filePrefix)))
	if err != nil {
		panic(err)
	}
	defer func(outputFile *os.File) {
		err := outputFile.Close()
		if err != nil {
			return
		}
	}(outputFile)

	// write sitemap to output file
	logger("i", "Writing textual sitemap...")
	err = printSitemap(basePage, 0, outputFile)
	if err != nil {
		return
	}

	logger("i", "Crawl finished!")
}

/*
Crawls one page:
 1. check if depth reached, and return if it has
 2. get all links on a page
 3. for each link
    . recursively set a new go routine running to crawl
*/
func crawl(page *Page, depth int, initialBaseURL *url.URL, globalWait *sync.WaitGroup) {
	logger("i", fmt.Sprintf("Starting crawl for: %s  ||  depth of %d", page.URL.String(), depth))
	// handle wait group at start
	defer globalWait.Done()

	// check depth and return if max depth
	if depth == 0 {
		return
	}

	// get all links on the given page
	newLinks, err := getLinksFromURL(page.URL, initialBaseURL)
	if err != nil {
		logger("e", fmt.Sprintf("Error getting links for: %s \n", page.URL.String()))
		return
	}

	for _, newL := range newLinks {
		// add to current pages Links slice
		page.links = append(page.links, newL)
		// only proceed if next depth > 0
		if depth-1 > 0 {
			// Crawl each page found on this page
			globalWait.Add(1)
			go crawl(newL, depth-1, initialBaseURL, globalWait)
		}
	}
}

func getLinksFromURL(link *url.URL, baseURL *url.URL) ([]*Page, error) {
	// get response
	resp, err := http.Get(link.String())
	if err != nil {
		logger("e", fmt.Sprintf("Error getting response from: %s", link.String()))
		return nil, err
	}

	// parse for <a> tags
	var newPages []*Page
	z := html.NewTokenizer(resp.Body)

	// used to avoid duplicate links being returned
	seen := make(map[string]bool)

	for {
		token := z.Next()

		switch {
		case token == html.ErrorToken:
			// End of page, return
			return newPages, nil
		case token == html.StartTagToken:
			// check if anchor tag found
			tag := z.Token()
			// worth noting here that this does not guarantee 100% of links, could be stuff in javascript somewhere
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

							// check if link has been seen on this page already
							if _, ok := seen[l.String()]; !ok {
								//link has not been seen before and is of the right domain
								lPage := &Page{URL: l}
								newPages = append(newPages, lPage)
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

// Takes a url string and returns an url.URL type if valid
func parseURL(URLString string) (*url.URL, error) {
	resultURL, err := url.Parse(URLString)
	if err != nil {
		logger("e", fmt.Sprintf("Could not parse URL: %s", URLString))
	}
	return resultURL, err
}

// Prints all links in the basePage with the given indent level
// To be used recursively
func printSitemap(basePage *Page, indent int, outputFile *os.File) error {
	indentString := "        "

	// print basePage URL
	if indent == 0 {
		_, err := io.WriteString(outputFile, fmt.Sprintf("%s\n", basePage.URL.String()))
		if err != nil {
			logger("e", fmt.Sprintf("Failed to write to file for: %s", basePage.URL.String()))
		}
	}

	// print basePage links recursively
	for _, l := range basePage.links {
		_, err := io.WriteString(outputFile, fmt.Sprintf("%s - %s\n", strings.Repeat(indentString, indent), l.URL.String()))
		if err != nil {
			logger("e", fmt.Sprintf("Failed to write to file for: %s", l.URL.String()))
		}
		// print children links
		err = printSitemap(l, indent+1, outputFile)
		if err != nil {
			return err
		}
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

// create a directory given in a path if one doesn't already exist
func createDirIfNotExist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// generate a filename using a given prefix and time.now
func generateDateFileName(datePrefix string) string {
	now := time.Now().UTC().Format("2006-01-02_15-04-05")
	return fmt.Sprintf("%s_%s", datePrefix, now)
}
