# Web page scraper
Concurrent scraper for a specific domain.

**Installation**

 1. `go get`


**Example usage**

    go run crawler -d 5 -b monzo.com

    go run crawler --help

**Output**
| \<domain\>.txt | textual sitemap text file |
|--|--|
| run.log |rolling log file  |
		

**Tests**
	
	

    go test

