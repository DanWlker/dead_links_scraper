/*
Copyright © 2024 DanWlker danielhee2@gmail.com
*/
package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"golang.org/x/net/html"
)

var domainNotFoundError = errors.New("Domain not found")

var rootCmd = &cobra.Command{
	Use:   "dead_links_scraper",
	Short: "Scrape dead links on a website",
	Run: func(cmd *cobra.Command, args []string) {
		parallelFlag, err := cmd.Flags().GetBool("parallel")
		if err != nil {
			cobra.CheckErr(fmt.Errorf("cmd.Flags().GetBool: %w", err))
		}

		if len(args) < 1 {
			cobra.CheckErr(fmt.Errorf("Please provide the url domain"))
		}

		startFlag, err := cmd.Flags().GetString("start")
		if err != nil {
			cobra.CheckErr(fmt.Errorf("cmd.Flags().GetBool: %w", err))
		}

		rootRun(parallelFlag, startFlag, args[0])
	},
}

func rootRun(parallel bool, start, domain string) {
	fmt.Printf("Base domain: %s\n", domain)

	if parallel {
		return
	}

	if start != "" {
		fmt.Println("Starting search from: " + start)
	}

	found, dead := make(map[string]bool), make(map[string]string)

	startDomain, err := url.JoinPath(domain, start)
	if err != nil {
		cobra.CheckErr(fmt.Errorf("url.JoinPath: %w", err))
	}
	recursiveScrape(startDomain, found, dead, domain)

	writer := tabwriter.NewWriter(
		os.Stdout, 0, 2, 4, ' ', 0,
	)

	_, _ = writer.Write([]byte("\nPage\tLink\n"))
	for fullLink, page := range dead {
		_, link := path.Split(fullLink)
		_, _ = writer.Write(
			[]byte(page + "\t" + link + "\n"),
		)
	}

	writer.Flush()
}

func recursiveScrape(domain string, found map[string]bool, dead map[string]string, baseDomain string) error {
	if strings.HasPrefix(domain, "/") {
		var err error
		domain, err = url.JoinPath(baseDomain, domain)
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("\nScraping ", domain)

	// save as checked page
	if found[domain] {
		fmt.Println("Visited, returning")
		return nil
	}
	found[domain] = true

	// fetch the page
	fmt.Printf("Fetching %v\n", domain)
	resp, err := http.Get(domain)
	// check if fails, save it and return
	if err != nil {
		fmt.Println(err)
		return domainNotFoundError
	}

	if resp.StatusCode != 200 {
		fmt.Println("Call is successful but server returned " + strconv.Itoa(resp.StatusCode))
		return domainNotFoundError
	}

	if resp.Request.URL.String() != domain {
		fmt.Printf("Redirected to %v, adding it to found\n", resp.Request.URL)
		if found[resp.Request.URL.String()] {
			fmt.Println("Visited, returning")
			return nil
		}
		found[resp.Request.URL.String()] = true
	}

	// do not continue check if its in a different domain
	if !strings.HasPrefix(domain, baseDomain) {
		return nil
	}

	// if succeeded, serialize the html, and grab all the links
	z := html.NewTokenizer(resp.Body)
	var links []string

Loop:
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			break Loop
		case html.StartTagToken, html.EndTagToken:
			tn, hasAttr := z.TagName()
			if len(tn) == 1 && tn[0] == 'a' && hasAttr {
				for ok := true; ok; {
					key, val, hasAttr := z.TagAttr()
					ok = hasAttr
					if string(key) == "href" {
						links = append(links, string(val))
						break
					}
				}
			}
		}

	}

	for _, link := range links {
		if err := recursiveScrape(link, found, dead, baseDomain); err != nil {
			dead[link] = domain
		}
	}

	return nil
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("parallel", "p", false, "Run scraper concurrently")
	rootCmd.Flags().StringP("start", "s", "", "Defines which relative path from the base domain to start searching from. Ex: /believe")
}
