package main

import (
	"flag"
	"fmt"
	"github.com/gocolly/colly/v2"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
)

func main() {
	var startURL = flag.String("url", "", "Start URL")
	var dest = flag.String("dest", "", "Destination folder")
	flag.Parse()

	if len(*startURL) == 0 || len(*dest) == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	err := checkURL(*startURL)
	if err != nil {
		fmt.Errorf("%s", err)
	}
	dst, err := checkDest(*dest)
	if err != nil {
		fmt.Errorf("%s", err)

	}
	crawl(
		*startURL,
		dst,
	)
}

func checkURL(startUrl string) error {
	_, err := url.ParseRequestURI(startUrl)
	if err != nil {
		return err
	}
	return nil
}

func checkDest(dest string) (string, error) {
	dst, err := filepath.Abs(dest)
	if err != nil {
		return "", err
	}
	err = os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		return "", err
	}

	return dst, nil
}

func crawl(startURL string, dest string) {
	c := colly.NewCollector(
		colly.CacheDir(dest +"/nl_cache"),
		colly.URLFilters(
			regexp.MustCompile(fmt.Sprintf("%s.*", regexp.QuoteMeta(startURL))),
		),
		// sets the recursion depth for links to visit, goes on forever if not set
		colly.MaxDepth(5),
		// enables asynchronous network requests
		colly.Async(true),
		)
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})
	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// Extract the link from the anchor HTML element
		link := e.Attr("href")
		// Tell the collector to visit the link
		c.Visit(e.Request.AbsoluteURL(link))
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		path := r.Request.URL.Path
		HTMLfilename := r.Request.URL.RawQuery + ".html"
		if len(r.Request.URL.RawQuery) == 0 {
			HTMLfilename = "index.html"
		}
		absPath := filepath.Join(dest , path)
		_ = os.MkdirAll(absPath, os.ModePerm)
		absPathFile := filepath.Join(absPath, HTMLfilename)
		fmt.Println("Saving file", absPathFile)
		err := ioutil.WriteFile(absPathFile, r.Body, 0644)
		if err != nil {
			fmt.Println(err)
		}
	})

	err := c.Visit(startURL)
	if err != nil {
		fmt.Println(err)
	}
	c.Wait()
}
