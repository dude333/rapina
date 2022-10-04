# 𝚛𝚊𝚙𝚒𝚗𝚊

Download and process Brazilian companies' financial data directly from [CVM](http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/DFP/). [[Em português](./README.md)]

[![GitHub release](https://img.shields.io/github/tag/dude333/rapina.svg?label=latest)](https://github.com/dude333/rapina/releases)
[![Travis](https://img.shields.io/travis/dude333/rapina/master.svg)](https://travis-ci.org/dude333/rapina)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

# 1. Installation

No installation required, just download the [latest released executable](https://github.com/dude333/rapina/releases). Then open a terminal ([CMD](https://superuser.com/a/340051/61616) on Windows) and run the commands shown below.

# 2. Commands

For the first time, run the following command:

    ./rapina get

Then, to get a company report, together with a summary for the companies from the same sector:

    ./rapina report <company>

## 2.1. `get`| Download and store financial data into the local database

    ./rapina get [-s]

It downloads all files from CVM web server, parses their contents and stores on a sqlite database at `.data/rapina.db`.

This command must be run **at least once** before you run the other commands.

### 2.1.1 Option

```
  -s, --sectors   Download and sector classification for companies listed at B3
```

Used to get only a summary for the other companies from the same sector.

[![asciicast](https://asciinema.org/a/656x2hrtCFFZLVLa9fGGcetw7.svg)](https://asciinema.org/a/656x2hrtCFFZLVLa9fGGcetw7?speed=4&autoplay=1&loop=1)

## 2.2. `list`| List all companies

    ./rapina list

[![asciicast](https://asciinema.org/a/TbJyGaOodJUxEzjDySQu3MaEW.svg)](https://asciinema.org/a/TbJyGaOodJUxEzjDySQu3MaEW?autoplay=1&loop=1)

## 2.3. `report`| Create a spreadsheet with a company financial data

    ./rapina report [flags] company_name

A spreadsheet with the financial data will be created and, on another sheet, the summary of all companies in the same sector.

The sector list is obtained from B3 and saved in the `sector.yml` file (via `get -s` command). If you want to change the sector grouping, just edit this file.

### 2.3.1. Options

```
  -d, --outputDir string   Output directory [default: ./reports]
  -s, --scriptMode         Does not show companies list; uses the most similar
                           company name
```

On **Linux** or **macOS**, use the arrow keys to navigate through the companies list. On **Windows**, use <kbd>j</kbd> and <kbd>k</kbd>.

[![asciicast](https://asciinema.org/a/jhmHxzgROtc8EBh3tkSwYTaa9.svg)](https://asciinema.org/a/jhmHxzgROtc8EBh3tkSwYTaa9?autoplay=1&loop=1)

### 2.3.2. Examples

    ./rapina report WEG

The spreadsheet will be saved at `./reports`

    ./rapina report "TEC TOY" -s -d /tmp/output

The spreadsheet will be saved at `/tmp/output`

# 3. Troubleshooting

Some Linux distributions (e.g. Fedora 34) might face some issues regarding the signer authority (Global Sign) that B3 is using on its SSL certificates. In case of `x509: certificate signed by unknown authority` error, one should manually import the Root CA certificate into the O.S. trusted database:

**Fedora 34 / CentOS** 

1. Download the Issuer Root Cert

    `curl http://secure.globalsign.com/cacert/gsrsaovsslca2018.crt > /tmp/global-signer.der`

2. Convert from .der to .pem

    `openssl x509 -inform der -in /tmp/global-signer.der -out /tmp/globalsignroot.pem`

3. Move the .pem file to the anchors folder

    `sudo cp /tmp/globalsignroot.pem /usr/share/pki/ca-trust-source/anchors/`

4. Update the trusted certificates database

    `sudo update-ca-trust`

**Ubuntu** 

1. Download the Issuer Root Cert

    `curl https://secure.globalsign.net/cacert/Root-R1.crt > /tmp/GlobalSign_Root_CA.crt`
    `curl https://secure.globalsign.net/cacert/Root-R2.crt > /tmp/GlobalSign_Root_CA_R2.crt`

2. Move the .crt files to the certificates folder

    `sudo cp /tmp/GlobalSign_Root_CA.crt /usr/local/share/ca-certificates/`
    `sudo cp /tmp/GlobalSign_Root_CA_R2.crt /usr/local/share/ca-certificates/`

3. Update the trusted certificates database

    `sudo update-ca-trust`


# 4. How to compile

If you want to compile your own executable, you need first to [download and install](https://golang.org/dl/) the Go compiler. Then follow these steps:

1. `go get github.com/dude333/rapina`
2. `cd $GOPATH/src/github.com/dude333/rapina`
3. Change to the cli directory (`cd cli`)
4. Compile using the Makefile (`make`). _To cross compile for Windows on Linux, use `make win`_.

# 5. Contributing

1. Fork it
2. `cd $GOPATH/src/github.com/your_username`
3. Download your fork to your PC (`git clone https://github.com/your_username/rapina && cd rapina`)
4. Create your feature branch (`git checkout -b my-new-feature`)
5. Make changes and add them (`git add .`)
6. Commit your changes (`git commit -m 'Add some feature'`)
7. Push to the branch (`git push origin my-new-feature`)
8. Create new pull request

# 6. Screenshot

![WEG](https://i.imgur.com/czPhPkH.png)

# 7. License

MIT
