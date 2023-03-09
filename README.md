[![Go Test](https://github.com/james-ecd/sitemapper/actions/workflows/go-tests.yml/badge.svg)](https://github.com/james-ecd/sitemapper/actions/workflows/go-tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/james-ecd/sitemapper)](https://goreportcard.com/report/github.com/james-ecd/sitemapper)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)


# Sitemapper
Tool for crawling a given domain, and mapping out it's URL's based on scraped links.
A play project for learning go.

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
