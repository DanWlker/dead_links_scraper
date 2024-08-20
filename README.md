Web Scraper for dead links in a website that supports both parallel and sequential execution

## To Build:

```go
go build .
```

## Usage:

```sh
./dead_links_scraper https://<your_link>
```

To start from a relative path, specify `-s`

```sh
./dead_links_scraper -s /relative_path https://<your_link>
```

To execute in parallel, specify `-p`

```sh
./dead_links_scraper -p https://<your_link>
```

