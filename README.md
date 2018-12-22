# 𝚛𝚊𝚙𝚒𝚗𝚊

Download and process Brazilian companies' financial data directly from CVM web server:

    http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/DFP/

[![GitHub release](https://img.shields.io/github/tag/dude333/rapina.svg?label=latest)](https://github.com/dude333/rapina/releases)
[![Travis](https://img.shields.io/travis/dude333/rapina/master.svg)](https://travis-ci.org/dude333/rapina)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

# Commands

## Download and store financial data into the local database

    ./rapina get

It downloads all files from CVM web server, parses their contents and stores on a sqlite database at `.data/rapina.db`.

This command must be run **at least once** before you run the `report`.

## List all companies

    ./rapina list

## Create a spreadsheet with a company financial data

    ./rapina report [flags] company_name

### Options

```
  -d, --outputDir string   Output directory
  -s, --scriptMode         Does not show companies list; uses the most similar
                           company name [default: ./reports]
```

On **Linux** or **macOS**, use the arrow keys to navigate through the companies list. On **Windows**, use <kbd>j</kbd> and <kbd>k</kbd>.

### Examples

    ./rapina report WEG

The spreadsheet will be saved at `./reports`

    ./rapina report "TEC TOY" -s -d /tmp/output

The spreadsheet will be saved at `/tmp/output`

# How to compile

1. Clone this repo to your PC (`git clone https://github.com/dude333/rapina`)
2. Change to CLI directory (`cd rapina/cli`)
3. Compile using the Makefile (`make`). To _cross compile_ for Windows on Linux, use `make win`.

# Contributing

1. Fork it
2. Download your fork to your PC (`git clone https://github.com/your_username/rapina && cd rapina`)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Make changes and add them (`git add .`)
5. Commit your changes (`git commit -m 'Add some feature'`)
6. Push to the branch (`git push origin my-new-feature`)
7. Create new pull request

# License

MIT
