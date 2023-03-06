# Sitemapper
Tool for crawling a given domain, and mapping out it's URL's based on scraped links

**Installation**

 1. `go get`


**Example usage**

    go run crawler -d 5 -b example.com

    go run crawler --help

**Output**
- `.txt`: textual sitemap text file |
- `.log`: log file 
		

**Tests**

    go test
