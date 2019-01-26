# 𝚛𝚊𝚙𝚒𝚗𝚊

Download e processamento de dados financeiros de empresas brasileiras diretamente da [CVM](http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/DFP/). [[In English](./README_en.md)]

[![GitHub release](https://img.shields.io/github/tag/dude333/rapina.svg?label=latest)](https://github.com/dude333/rapina/releases)
[![Travis](https://img.shields.io/travis/dude333/rapina/master.svg)](https://travis-ci.org/dude333/rapina)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

# 1. Instalação

Não é necessário instalar, basta baixar o executável da [página de release](https://github.com/dude333/rapina/releases).

Abra o terminal ([CMD](https://superuser.com/a/340051/61616) no Windows) e rode os comandos listados abaixo.

# 2. Comandos

Na primeira vez, rodar o seguinte comando:

    ./rapina get

Depois, para obter o relatório de uma determinada empresa, com o resumo das empresas do mesmo setor:

    ./rapina report <empresa>

## 2.1. `get`| Download e armazenamento de dados financeiros no banco de dados local

    ./rapina get [-s]

Baixa todos os arquivos disponíveis no servidor da CVM, processa o conteúdo e o armazena num banco de dados sqlite em `.data/rapina.db`.

Este comando deve ser executado **pelo menos uma vez** antes dos outros comandos.

### 2.1.1 Opção

```
  -s, --sectors   Baixa a classificação setorial das empresas e fundos negociados na B3
```

Usado para obter apenas o resumo dos indicadores das empresas do mesmo setor.

[![asciicast](https://asciinema.org/a/656x2hrtCFFZLVLa9fGGcetw7.svg)](https://asciinema.org/a/656x2hrtCFFZLVLa9fGGcetw7?speed=4&autoplay=1&loop=1)

## 2.2. `list`| Lista todas as empresas disponíveis

    ./rapina list

[![asciicast](https://asciinema.org/a/TbJyGaOodJUxEzjDySQu3MaEW.svg)](https://asciinema.org/a/TbJyGaOodJUxEzjDySQu3MaEW?autoplay=1&loop=1)

## 2.3. `report`| Cria uma planilha com os dados financeiros de uma empresa

    ./rapina report [flags] empresa

Será criada uma planilha com os dados financeiros (BP, DRE, DFC) e, em outra aba, o resumo de todas as empresas do mesmo setor.

A lista setorial é obtida da B3 e salva no arquivo `setor.yml` (via comando `get -s`). Caso deseje alterar o agrupamento setorial, basta editar este arquivo.

### 2.3.1. Opções

```
  -d, --outputDir string   Diretório onde a planilha será salva
                           [default: ./reports]
  -s, --scriptMode         Não lista as empresas; usa a com nome mais próximo
```

No **Linux** ou **macOS**, use as setas para navegar na lista das empresas. No **Windows**, use <kbd>j</kbd> e <kbd>k</kbd>.

[![asciicast](https://asciinema.org/a/jhmHxzgROtc8EBh3tkSwYTaa9.svg)](https://asciinema.org/a/jhmHxzgROtc8EBh3tkSwYTaa9?autoplay=1&loop=1)

### 2.3.2. Exemplos

    ./rapina report WEG

A planilha será salva em `./reports`

    ./rapina report "TEC TOY" -s -d /tmp/output

A planilha será salva em `/tmp/output`

# 3. Como compilar

Se quiser compilar seu próprio executável, primeiro [baixe e instale](https://golang.org/dl/) o compilador Go. Depois execute estes passos:

1. `go get github.com/dude333/rapina`
2. `cd $GOPATH/src/github.com/dude333/rapina`
3. Change to the cli directory (`cd cli`)
4. Compile using the Makefile (`make`). _To cross compile for Windows on Linux, use `make win`_.

# 4. Contribua

1. Faça um fork deste projeto
2. `cd $GOPATH/src/github.com/your_username`
3. `git clone https://github.com/your_username/rapina && cd rapina`
4. `git checkout -b my-new-feature`
5. `git add .`
6. `git commit -m 'Add some feature'`
7. `git push origin my-new-feature`
8. Crie um _pull request_

# 5. Screenshot

![WEG](https://i.imgur.com/czPhPkH.png)

# 6. License

MIT
