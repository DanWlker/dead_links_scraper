/*
Copyright © 2024 DanWlker danielhee2@gmail.com
*/
package cmd

import (
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

		rootRun(parallelFlag, args[0])
	},
}

func rootRun(parallel bool, domain string) {
	fmt.Printf("Base domain: %s\n", domain)

	if parallel {
		return
	}

	found, dead := make(map[string]bool), make(map[string]bool)

	recursiveScrape(domain, found, dead, domain)

	writer := tabwriter.NewWriter(
		os.Stdout, 0, 2, 4, ' ', 0,
	)

	_, _ = writer.Write([]byte("\nPage\tLink\n"))
	for link := range dead {
		page, link := path.Split(link)
		_, _ = writer.Write(
			[]byte(page + "\t" + link + "\n"),
		)
	}

	writer.Flush()
}

func recursiveScrape(domain string, found, dead map[string]bool, baseDomain string) {
	if strings.HasPrefix(domain, "/") {
		var err error
		domain, err = url.JoinPath(baseDomain, domain)
		if err != nil {
			fmt.Println(err)
		}
	}

	// save as checked page
	if found[domain] {
		return
	}
	found[domain] = true

	// fetch the page
	fmt.Printf("Fetching %v\n", domain)
	resp, err := http.Get(domain)
	// check if fails, save it and return
	if err != nil {
		fmt.Println(err)
		dead[domain] = true
	}

	if resp.StatusCode != 200 {
		fmt.Println("Call is successful but server returned " + strconv.Itoa(resp.StatusCode))
		dead[domain] = true
	}

	// do not continue check if its in a different domain
	if !strings.HasPrefix(domain, baseDomain) {
		return
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
		recursiveScrape(link, found, dead, baseDomain)
	}
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("parallel", "p", false, "Run scraper concurrently")
}
