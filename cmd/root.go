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
	"sync"
	"text/tabwriter"

	"dead_links_scraper/pkg"

	"github.com/spf13/cobra"
	"golang.org/x/net/html"
)

var domainDeadError = errors.New("Domain dead error, domain accessed is dead")

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

	start, err := url.JoinPath(domain, start)
	if err != nil {
		cobra.CheckErr(fmt.Errorf("url.JoinPath: %w", err))
	}
	fmt.Println("Starting search from: " + start)
	var dead map[string]string

	if parallel {
		found, deadTemp := pkg.NewAtomicSet[string](), pkg.NewAtomicMap[string, string]()

		checkSetFound := func(s string) bool {
			insertSuccess := found.Insert(s)
			return !insertSuccess
		}

		var onScrapedPage func(s []string)
		onScrapedPage = func(s []string) {
			var wg sync.WaitGroup
			for _, link := range s {
				wg.Add(1)
				go func() {
					if err := scrape(
						link,
						checkSetFound,
						domain,
						wg.Done,
						onScrapedPage,
					); err != nil {
						deadTemp.Set(link, domain)
					}
				}()
			}

			wg.Wait()
		}

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			if err := scrape(
				start,
				checkSetFound,
				domain,
				wg.Done,
				onScrapedPage,
			); err != nil {
				cobra.CheckErr(fmt.Errorf("parallel scrape: %w", err))
			}
		}()
		wg.Wait()

		dead = deadTemp.ToMap()
	} else {
		found := make(map[string]bool)
		dead = make(map[string]string)

		checkSetFound := func(s string) bool {
			res := found[s]
			found[s] = true
			return res
		}

		var onScrapedPage func([]string)
		onScrapedPage = func(s []string) {
			for _, link := range s {
				if err := scrape(
					link,
					checkSetFound,
					domain,
					func() {},
					onScrapedPage,
				); err != nil {
					dead[link] = domain
				}
			}
		}

		if err := scrape(
			start,
			checkSetFound,
			domain,
			func() {},
			onScrapedPage,
		); err != nil {
			cobra.CheckErr(fmt.Errorf("sequential scrape: %w", err))
		}
	}

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

func scrape(
	domain string,
	checkSetFound func(string) bool,
	baseDomain string,
	cleanup func(),
	onScrapedPage func([]string),
) error {
	defer cleanup()

	if strings.HasPrefix(domain, "/") {
		var err error
		domain, err = url.JoinPath(baseDomain, domain)
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("\nScraping ", domain)

	if checkSetFound(domain) {
		fmt.Println("Visited, returning")
		return nil
	}

	fmt.Printf("Fetching %v\n", domain)
	resp, err := http.Get(domain)
	// check if fails, save it and return
	if err != nil {
		fmt.Println(err)
		return domainDeadError
	}

	if resp.StatusCode != 200 {
		fmt.Println("Call is successful but server returned " + strconv.Itoa(resp.StatusCode))
		return domainDeadError
	}

	if resp.Request.URL.String() != domain {
		fmt.Printf("Redirected to %v, adding it to found\n", resp.Request.URL)
		if checkSetFound(resp.Request.URL.String()) {
			fmt.Println("Visited, returning")
			return nil
		}
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

			if len(tn) != 1 || tn[0] != 'a' || !hasAttr {
				continue Loop
			}

			for ok := true; ok; {
				key, val, hasAttr := z.TagAttr()
				ok = hasAttr

				if string(key) != "href" {
					continue
				}

				links = append(links, string(val))
				break
			}
		}

	}

	onScrapedPage(links)
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
