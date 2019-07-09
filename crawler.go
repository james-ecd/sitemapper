/* Task:
 simple web crawler
 Given a URL, it should output a simple textual sitemap, showing the links between pages.
- The crawler should be limited to one subdomain
- not follow external links

show links 
 */

/* Method:

	- struct to contain a pages url and sub-pages
	- main func sets recursive crawler off on base domain with given depth
	- walk function does the following:
		1) check if depth == 0, and return if true
		2) get links in page
		3) for each link, set recursive call to parse its link with decremented depth
		4) send link struct pointer to recursive call so when they all fold back down the walked structure is preserved in struct
	- after all walking routines finished, output structure to text file

*/

package main

import (
	"log"
	"os"
	"fmt"
	"flag"
	"net/url"
	"sync"
)

// Link struct represents a page and all links found on the page
type Link struct {
	url *url.URL
	links []*Link
}

// Concurrency safe set implementation of seen links. saves us from traversing previously traversed links
type Processed struct {
	links map[string]bool
	mux sync.Mutex
}

var processed Processed
var global_wait sync.WaitGroup

func main() {
	// Setup logger
	f, err := os.OpenFile("run.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// Define and save command line args
	base_url := flag.String("b", "monzo.com", "Starting URL to crawl from")
	search_depth := flag.Int("d", 7, "Number of levels you want to traverse (depth)")
	flag.Parse()

	// Start the crawl
	log.Print(fmt.Sprintf("------- STARTING NEW CRAWL FOR: %s -------", *base_url))
	
}

// Allows for levels of severity in our logger
func logger(severity string, message string) {
	switch severity {
		case "error":
			log.Print(fmt.Sprintf("[ERROR] %s", message))
		case "info":
			log.Print(fmt.Sprintf("[INFO] %s", message))
	}
}
