# gomls

[![License:
MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/TsekNet/gomls/blob/master/LICENSE)

Command-line tool that scrapes real estate listings from the web (with filters).

## Installation

Download the [executable](https://github.com/TsekNet/gomls/releases) from `/releases` (right side menu).

## Usage

```sh
.\gomls.exe -help
Usage: .\gomls.exe <flags> <subcommand> <subcommand args>

Subcommands:
        commands         list all command names
        flags            describe all known top-level flags
        help             describe subcommands and their syntax

Subcommands for Lists items with optional output format:
        list             Lists items with optional output format


Top-level flags (use ".\gomls.exe flags" for a full list):
  -verbose=true: print info level logs to stdout
  ```

```sh
.\gomls.exe list -help
gomls.exe list
  -baths int
        Filter by number of baths
  -beds int
        Filter by number of beds
  -location string
        Filter by location of the properties (can be neighborhood, zip code, etc.).
        Type this into your search bar on zillow.com if you want to confirm the format. (default "10001")
  -output string
        Output format
        Must be one of: [plain, table, json, html, csv] (default "plain")
  -price int
        Filter by price
  -property_type string
        Filter by property type
        Must be one of: [APARTMENT, CONDO, MULTI_FAMILY, SINGLE_FAMILY]
  -sold
        Filter by sold properties
```

## Screenshots

### CSV output, sold listings

![CSV, no filter](media/csv.png)

### HTML output, all filters

![HTML with filter](media/html.png)

## Background

As a side project, mostly to get more comfortable with Go I aimed to build a tool that would scrape zillow.com for recent listings and sales. I aimed to allow flexible output types via a cross-platform (Linux, Mac, Windows) binary that can be shipped to friends to run on their systems. I also wanted to learn more about web scraping in Go (as opposed to using Python's `beautifulsoup` library). I didn't find any free, open source web scrapers for real estate data when searching online. All the ones I found were either a Chrome extension, or a paid API.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.
