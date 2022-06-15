# 𝚛𝚊𝚙𝚒𝚗𝚊

Download e processamento de dados<sup>[1](#disclaimer)</sup> financeiros de empresas brasileiras diretamente da [CVM](http://dados.cvm.gov.br/dados/CIA_ABERTA/DOC/DFP/).

[![GitHub release](https://img.shields.io/github/tag/dude333/rapina.svg?label=latest)](https://github.com/dude333/rapina/releases)
[![Travis](https://img.shields.io/travis/dude333/rapina/master.svg)](https://travis-ci.org/dude333/rapina)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

Este programa baixa e processa os arquivos CSV do site da CVM e os armazena em um banco de dados local (sqlite), onde são extraídos os dados **consolidados** do balanço patrimonial, fluxo de caixa, DRE (demonstração de resultado), DVA (demonstração de valor adicionado).

São coletados vários arquivos CSV desde 2010. Cada um destes arquivos contém informações do ano corrente e também do ano anterior, dessa forma foi possível extrair também os dados de 2009.

Com base nestes dados, são criados os relatórios por empresa, com um comparativo de outras empresas do mesmo setor. A classificação dos setores é baixada do site da Bovespa e armazenada no arquivo setores.yml (no formato [YAML](https://medium.com/@akio.miyake/introdu%C3%A7%C3%A3o-b%C3%A1sica-ao-yaml-para-ansiosos-2ac4f91a4443)), que pode ser editado para se adequar aos seus critérios, caso necessário.

A partir do release v0.11.0, passou-se a usar os dados trimestrais para compor os valores do ano corrente, usando-se para isso os últimos 4 trimestre ([TTM](#ttm-calc)), ou seja, a soma dos dados trimestrais do ano corrente com alguns do ano anterior, mantendo-se assim uma mesma base de comparação com os anos anteriores. 


# 1. Instalação

Não é necessário instalar, basta baixar o executável da [página de release](https://github.com/dude333/rapina/releases) e renomeie o executável para `rapina.exe` (no caso do Windows) ou `rapina` (para o Linux ou macOS).

Abra o terminal ([CMD](https://superuser.com/a/340051/61616) no Windows) e rode os comandos listados abaixo.

# 2. Uso

Na primeira vez, rodar o seguinte comando para baixar e processar os arquivos do site da CVM:

    ./rapina update

Depois, para obter o relatório de uma determinada empresa, com o resumo das empresas do mesmo setor:

    ./rapina report <empresa>

_Eventualmente, as empresas corrigem algum dado e enviam um novo arquivo à CVM, então é recomendável rodar o `rapina update` periodicamente._

# 3. Detalhe dos Comandos

## 3.1. update

**Download e armazenamento de dados financeiros no banco de dados local.**

    ./rapina update [-s]

Baixa todos os arquivos disponíveis no servidor da CVM, processa o conteúdo e o armazena num banco de dados sqlite em `.data/rapina.db`.

Este comando deve ser executado **pelo menos uma vez** antes dos outros comandos.

### 3.1.1 Opção

```
  -s, --sectors   Baixa a classificação setorial das empresas e fundos negociados na B3
```

Usado para obter apenas o arquivo de classificação setorial atualizado.

## 3.2. list

**Listas**

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

### 3.2.3 Lista empresas com critério de lucro líquido

```
  -l, --lucroLiquido número   Lista empresas com lucros lucros positivos e com a taxa de crescimento definida
```

Lista as empresas com lucros líquidos positivos e com uma taxa de crescimento definida em relação ao mês anterior. 
Por exemplo:
* Para listar as empresas com crescimento mínimo de 10% em relação ao ano anterior: `./rapina list -l 0.1`
* Para listar as empresas com variação no lucro de maiores que -5% em relação ao ano anterior: `./rapina list -l -0.05`


## 3.3. report

**Cria uma planilha com os dados financeiros de uma empresa.**

    ./rapina report [opções] empresa

Será criada uma planilha com os dados financeiros (BP, DRE, DFC) e, em outra aba, o resumo de todas as empresas do mesmo setor.

A lista setorial é obtida da B3 e salva no arquivo `setor.yml` (via comando `update -s`). Caso deseje alterar o agrupamento setorial, basta editar este arquivo. Mas lembre-se que ao rodar o `update -s` o arquivo será sobrescrito.

No **Linux** ou **macOS**, use as setas para navegar na lista das empresas. No **Windows**, use <kbd>j</kbd> e <kbd>k</kbd>.

### 3.3.1. Opções

```
  -a, --all                Mostra todos os indicadores
  -x, --extraRatios        Reporte de índices extras
  -F, --fleuriet           Capital de giro no modelo Fleuriet
  -o, --omitSector         Omite o relatório das empresas do mesmo setor
  -d, --outputDir string   Diretório onde o relatório será salvo (default "reports")
  -s, --scriptMode         Para modo script (escolhe a empresa com nome mais próximo)
  -f, --showShares         Mostra o número de ações e free float

```


### 3.3.2. Exemplos

    ./rapina report WEG

A planilha será salva em `./reports`

    ./rapina report "TEC TOY" -s -d /tmp/output

A planilha será salva em `/tmp/output`

# 4. Nova funções

## 4.1. fii

**Relatórios relacionados aos Fundos de Investimento Imobiliários**

### 4.1.1. rendimentos

    ./rapina fii rendimentos [-n] ABCD11 EFGH11...

Onde `-n` é o número de meses a serem apresentados.

E como parâmetros, passe uma lista de FIIs separados por espaço.

#### 4.1.1.1 Exemplo

    ./rapina fii rendimentos -n 2 knip11 hfof11

```
-------------------------------------------------------------------
KNIP11
-------------------------------------------------------------------
  DATA COM       RENDIMENTO     COTAÇÃO       YELD      YELD a.a.
  ----------     ----------     ----------    ------    ---------
  2021-04-30     R$    1,00     R$  113,00     0,88%       11,15%
  2021-03-31     R$    1,02     R$  115,95     0,88%       11,08%
-------------------------------------------------------------------
HFOF11
-------------------------------------------------------------------
  DATA COM       RENDIMENTO     COTAÇÃO       YELD      YELD a.a.
  ----------     ----------     ----------    ------    ---------
  2021-04-30     R$    0,60     R$   99,75     0,60%        7,46%
  2021-03-31     R$    0,56     R$  100,70     0,56%        6,88%
-------------------------------------------------------------------

```

# 4.2. server

**Web server para visualização dos relatórios no browser**

## 4.2.1. Exemplo

    ./rapina server

    2021/05/11 19:23:15 Listening on :3000...

Para visualizar a página, abrir o link http://localhost:3000

**NOTA:** Por hora só está disponível o relatório de rendimentos de FIIs.


# 5. Possíveis problemas

Algumas distribuições Linux (Fedora 34, por exemplo) podem encontrar problemas com as autoridades certificadores (Global Sign) presentes nos certificados SSL dos websites da B3. Em caso de erro `x509: certificate signed by unknown authority`, deve-se importar manualmente o Root CA para o trusted database do sistemas operacional:

**Fedora 34 / CentOS** 

1. Realizar o download do Issuer Root Cert

    `curl http://secure.globalsign.com/cacert/gsrsaovsslca2018.crt > /tmp/global-signer.der`

2. Converter de .der para .pem

    `openssl x509 -inform der -in /tmp/global-signer.der -out /tmp/globalsignroot.pem`

3. Importar .pem arquivo para pasta de anchors

    `sudo cp /tmp/globalsignroot.pem /usr/share/pki/ca-trust-source/anchors/`

4. Atualizar base de trusted certificates

    `sudo update-ca-trust`

**Ubuntu** 

1. Realizar o download do Issuer Root Cert

    `curl https://secure.globalsign.net/cacert/Root-R1.crt > /tmp/GlobalSign_Root_CA.crt`
    `curl https://secure.globalsign.net/cacert/Root-R2.crt > /tmp/GlobalSign_Root_CA_R2.crt`

2. Importar .crt arquivos para pasta de certificados

    `sudo cp /tmp/GlobalSign_Root_CA.crt /usr/local/share/ca-certificates/`
    `sudo cp /tmp/GlobalSign_Root_CA_R2.crt /usr/local/share/ca-certificates/`

3. Atualizar base de trusted certificates

    `sudo update-ca-trust`

# 6. Como compilar

Se quiser compilar seu próprio executável, primeiro [baixe e instale](https://golang.org/dl/) o compilador Go (v1.16 ou maior). Depois execute estes passos:

1. `git clone github.com/dude333/rapina`
2. `cd rapina`
3. `make`

O executável será criado na pasta `bin`. Você pode movê-lo para outro local. Ao rodar a primeira vez, apenar o executável é necessário, mas após rodá-lo, será criado um diretório `.data` que deverá ser movido junto com o executável, caso queira trazer o dados.

IMPORTANTE: para compilar a biblioteca do sqlite, é necessário ter um compilador C instalado na máquina (para o Windows, mais detalhes [aqui](https://github.com/mattn/go-sqlite3#windows)).

# 7. Contribua

1. Faça um fork deste projeto no [github.com](github.com/dude333/rapina)
2. `git clone https://github.com/`*your_username*`/rapina && cd rapina`
3. `git checkout -b `*my-new-feature*
4. Faça as modificações
5. `git add .`
6. `git commit -m 'Add some feature'`
7. `git push origin my-new-feature`
8. Crie um _pull request_

# 8. Screenshot

![WEG](https://i.imgur.com/czPhPkH.png)


# 9. License

MIT




<br />
<br />
<br />
<a name="disclaimer">1</a>: *Os dados são fornecidos "no estado em que se encontram" e somente para fins informativos, não para fins comerciais ou de consultoria.*
