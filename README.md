Web Scraper for dead links in a website. Refer [parallel](https://github.com/DanWlker/dead_links_scraper/tree/parallel) for concurrent version

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

## Caveats:

`-p` doesn't work currently, an example implementation is in the parallel branch

