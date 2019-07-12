package main

import (
	"bufio"
	"net/url"
	"os"
	"sync"
	"testing"
)

// test for parseURL
func TestParseURL(t *testing.T) {

	URLString := "monzo.com"

	resultsURL, err := parseURL(URLString)

	if err != nil {
		t.Errorf("There was an error parsing the url: %s", URLString)
	}

	if resultsURL.String() != URLString {
		t.Errorf("Parsed URL does not match input URL string...")
	}

}

// test for getLinksFromURL
func TestGetLinksFromURL(t *testing.T) {
	/*
		Issue with this test is finding a static webstie where links wont change.
		If we don't have such a website, than the test if liable for false negatives.
		In a production enviroment you would host a simple webpage specifically for this test,
		however to save time I have used a static resource found online, that hasn't change in over 10 years.
	*/

	testURLString := "https://www.tic.com"
	testURL, _ := url.Parse(testURLString)

	expectedResponses := map[string]bool{
		"https://www.tic.com/index.html":             false,
		"https://www.tic.com/bios/index.html":        false,
		"https://www.tic.com/books/index.html":       false,
		"https://www.tic.com/opensource/index.html":  false,
		"https://www.tic.com/partners/index.html":    false,
		"https://www.tic.com/rfcs/index.html":        false,
		"https://www.tic.com/whitepapers/index.html": false,
	}

	linksSlice, _ := getLinksFromURL(testURL, testURL)

	// Verify no unexpected links, and set all expected found links to true
	for _, l := range linksSlice {
		if _, ok := expectedResponses[l.URL.String()]; !ok {
			t.Errorf("Unexpected link found: %s", l.URL.String())
		} else {
			expectedResponses[l.URL.String()] = true
		}
	}

	// Verify every expected link was found
	for k, v := range expectedResponses {
		if !v {
			//link wasn't found
			t.Errorf("Expected URL: %s was not found...", k)
		}
	}
}

//test for main crawl function
func TestCrawl(t *testing.T) {
	// create dummy URL and LINK
	testURLString := "https://www.tic.com"
	testURL, _ := url.Parse(testURLString)
	baseLink := &Link{URL: testURL}

	// run a crawl of depth 1
	var wg sync.WaitGroup
	wg.Add(1)
	crawl(baseLink, 1, testURL, &wg)
	wg.Wait()

	// verify data structure in baseLink is correct
	expectedURLs := map[string]bool{
		"https://www.tic.com/index.html":             false,
		"https://www.tic.com/bios/index.html":        false,
		"https://www.tic.com/books/index.html":       false,
		"https://www.tic.com/opensource/index.html":  false,
		"https://www.tic.com/partners/index.html":    false,
		"https://www.tic.com/rfcs/index.html":        false,
		"https://www.tic.com/whitepapers/index.html": false,
	}

	// check there are no unexpected links and mark found expected ones
	for _, l := range baseLink.links {
		if _, ok := expectedURLs[l.URL.String()]; !ok {
			t.Errorf("Unexpected link found: %s", l.URL.String())
		} else {
			expectedURLs[l.URL.String()] = true
		}
	}
	// check all expected links were found
	for k, v := range expectedURLs {
		if !v {
			//link wasn't found
			t.Errorf("Expected URL: %s was not found...", k)
		}
	}
}

// test for print sitemap
func TestPrintSitemap(t *testing.T) {
	/*
		Worth noting there may be a better way of doing this by mocking the os
		and "writing" to a fake file. Decided to test it using a real file in
		the interest of time...
	*/
	// create file
	outputFile, _ := os.Create("test.txt")
	// create dummy links
	baseURL, _ := url.Parse("https://test.com")
	subURL1, _ := url.Parse("https://test.com/sub1/")
	subURL2, _ := url.Parse("https://test.com/sub2/")

	baseLink := &Link{URL: baseURL}
	subLink1 := &Link{URL: subURL1}
	subLink2 := &Link{URL: subURL2}

	baseLink.links = append(baseLink.links, subLink1, subLink2)

	// write to file
	printSitemap(baseLink, 0, outputFile)
	outputFile.Close()

	// verify file contents are correct
	file, _ := os.Open("test.txt")
	scanner := bufio.NewScanner(file)

	expectedStrings := map[int]string{
		0: "https://test.com",
		1: " - https://test.com/sub1/",
		2: " - https://test.com/sub2/",
	}

	index := 0
	for scanner.Scan() {
		line := scanner.Text()
		if line != expectedStrings[index] {
			t.Errorf("Line did not match: %s - %s", expectedStrings[index], line)
		}
		index++
	}
	file.Close()

	// delete file
	err := os.Remove("test.txt")
	if err != nil {
		t.Errorf("Couldnt delete file: %s", err)
	}
}
