# rapina

Download and process Brazilian companies' financial data directly from CVM web server:

    http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/DFP/

## Commands

### Download and store financial data into the local database

    ./rapina get

It downloads all files from CVM web server, parses their contents and stores on this sqlite database: `.data/rapina.db`.

### List all companies

    ./rapina list

### Create a spreadsheet with a company financial data

    ./rapina report <"COMPANY NAME">

For example:

    ./rapina report WEG
    ./rapina report "TEC TOY"

The spreadsheet will be saved at `.data/COMPANY_NAME.xlsx`

## How to compile

1. Clone this repo to your PC (`git clone https://github.com/dude333/rapina`)
2. Change to CLI directory (`cd rapina/cli`)
3. Compile using the Makefile (`make`). To cross compile for Windows on Linux, use `make win`.

## Contributing

1. Fork it
2. Download your fork to your PC (`git clone https://github.com/your_username/rapina && cd rapina`)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Make changes and add them (`git add .`)
5. Commit your changes (`git commit -m 'Add some feature'`)
6. Push to the branch (`git push origin my-new-feature`)
7. Create new pull request

## License

MIT
