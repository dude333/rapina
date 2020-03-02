# 𝚛𝚊𝚙𝚒𝚗𝚊

Download e processamento de dados financeiros de empresas brasileiras diretamente da [CVM](http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/DFP/).

[![GitHub release](https://img.shields.io/github/tag/dude333/rapina.svg?label=latest)](https://github.com/dude333/rapina/releases)
[![Travis](https://img.shields.io/travis/dude333/rapina/master.svg)](https://travis-ci.org/dude333/rapina)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

# 1. Instalação

Não é necessário instalar, basta baixar o executável da [página de release](https://github.com/dude333/rapina/releases).

Abra o terminal ([CMD](https://superuser.com/a/340051/61616) no Windows) e rode os comandos listados abaixo.

# 2. Uso

Na primeira vez, rodar o seguinte comando para baixar e processar os arquivos do site da CVM:

    ./rapina get

Depois, para obter o relatório de uma determinada empresa, com o resumo das empresas do mesmo setor:

    ./rapina report <empresa>

_Eventualmente, as empresas corrigem algum dado e enviam um novo arquivo à CVM, então é recomendável rodar o `rapina get` periodicamente._

# 3. Detalhe dos Comandos

## 3.1. get

**Download e armazenamento de dados financeiros no banco de dados local.**

    ./rapina get [-s]

Baixa todos os arquivos disponíveis no servidor da CVM, processa o conteúdo e o armazena num banco de dados sqlite em `.data/rapina.db`.

Este comando deve ser executado **pelo menos uma vez** antes dos outros comandos.

### 3.1.1 Opção

```
  -s, --sectors   Baixa a classificação setorial das empresas e fundos negociados na B3
```

Usado para obter apenas o resumo dos indicadores das empresas do mesmo setor.

## 3.2. list

**Listagens**

    ./rapina list

### 3.2.1 Lista todas as empresas disponíveis

```
  -e, --empresas               Lista todas as empresas disponíveis
```

### 3.2.2 Lista as empresas do mesmo setor

```
  -s, --setor string           Lista todas as empresas do mesmo setor
```

Por exemplo, para listar todas as empras do mesmo setor do Itaú: `./rapina lista -s itau`

O resultado mostra a lista das empresas do mesmo setor contidos no banco de dados e no arquivo **setores.yml**, que você pode editar caso queira realocar os setores das empresas.

### 3.2.3 Lista todas as empresas disponíveis

```
  -l, --lucroLiquido número   Lista empresas com lucros lucros positivos e com a taxa de crescimento definida
```

Lista as empresas com lucros líquidos positivos e com uma taxa de crescimento definida em relação ao mês anterior. 
Por exemplo:
* Para listar as empresas com crescimento mínimo de 10% em relação ao ano anterior: `./rapina list -l 0.1`
* Para listar as empresas com variação no lucro de pelo menos -5% em relação ao ano anterior: `./rapina list -l -0.05`


## 3.3. report

**Cria uma planilha com os dados financeiros de uma empresa.**

    ./rapina report [opções] empresa

Será criada uma planilha com os dados financeiros (BP, DRE, DFC) e, em outra aba, o resumo de todas as empresas do mesmo setor.

A lista setorial é obtida da B3 e salva no arquivo `setor.yml` (via comando `get -s`). Caso deseje alterar o agrupamento setorial, basta editar este arquivo. Mas lembre-se que ao rodar o `get -s` o arquivo será sobrescrito.

### 3.3.1. Opções

```
  -d, --outputDir string   Diretório onde a planilha será salva
                           [default: ./reports]
  -s, --scriptMode         Não lista as empresas; usa a com nome mais próximo
```

No **Linux** ou **macOS**, use as setas para navegar na lista das empresas. No **Windows**, use <kbd>j</kbd> e <kbd>k</kbd>.

### 3.3.2. Exemplos

    ./rapina report WEG

A planilha será salva em `./reports`

    ./rapina report "TEC TOY" -s -d /tmp/output

A planilha será salva em `/tmp/output`

# 4. Como compilar

Se quiser compilar seu próprio executável, primeiro [baixe e instale](https://golang.org/dl/) o compilador Go (v1.13 ou maior). Depois execute estes passos:

1. `git clone github.com/dude333/rapina`
2. Change to the cli directory (`cd rapina/cli`)
3. Compile using the Makefile (`make`). _To cross compile for Windows on Linux, use `make win`_.

# 5. Contribua

1. Faça um fork deste projeto no [github.com](github.com/dude333/rapina)
2. `git clone https://github.com/`*your_username*`/rapina && cd rapina`
3. `git checkout -b `*my-new-feature*
4. Faça as modificações
5. `git add .`
6. `git commit -m 'Add some feature'`
7. `git push origin my-new-feature`
8. Crie um _pull request_

# 6. Screenshot

![WEG](https://i.imgur.com/czPhPkH.png)

# 7. Screencasts

# 7.1 rapina get

[![asciicast](https://asciinema.org/a/656x2hrtCFFZLVLa9fGGcetw7.svg)](https://asciinema.org/a/656x2hrtCFFZLVLa9fGGcetw7?speed=4&autoplay=1&loop=1)

# 7.2 rapina list

[![asciicast](https://asciinema.org/a/TbJyGaOodJUxEzjDySQu3MaEW.svg)](https://asciinema.org/a/TbJyGaOodJUxEzjDySQu3MaEW?autoplay=1&loop=1)

# 7.3 rapina report

[![asciicast](https://asciinema.org/a/jhmHxzgROtc8EBh3tkSwYTaa9.svg)](https://asciinema.org/a/jhmHxzgROtc8EBh3tkSwYTaa9?autoplay=1&loop=1)

# 8. License

MIT
