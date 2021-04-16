package parsers

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

const (
	resJSON = `{"draw":2,"recordsFiltered":51,"recordsTotal":51,"data":[{"id":159719,"descricaoFundo":"BB PROGRESSIVO II FUNDO DE INVESTIMENTO IMOBILIÁRIO - FII","categoriaDocumento":"Aviso aos Cotistas - Estruturado","tipoDocumento":"Rendimentos e Amortizações","especieDocumento":"","dataReferencia":"31/03/2021","dataEntrega":"31/03/2021 17:36","status":"AC","descricaoStatus":"Ativo com visualização","analisado":"N","situacaoDocumento":"A","assuntos":null,"altaPrioridade":false,"formatoDataReferencia":"3","versao":1,"modalidade":"AP","descricaoModalidade":"Apresentação","nomePregao":"FII BB PRGII","informacoesAdicionais":"FII BB PRGII;","arquivoEstruturado":"","formatoEstruturaDocumento":null,"nomeAdministrador":null,"cnpjAdministrador":null,"cnpjFundo":null,"idTemplate":0,"idSelectNotificacaoConvenio":null,"idSelectItemConvenio":0,"indicadorFundoAtivoB3":false,"idEntidadeGerenciadora":null,"ofertaPublica":null,"numeroEmissao":null,"tipoPedido":null,"dda":null},{"id":151499,"descricaoFundo":"BB PROGRESSIVO II FUNDO DE INVESTIMENTO IMOBILIÁRIO - FII","categoriaDocumento":"Aviso aos Cotistas - Estruturado","tipoDocumento":"Rendimentos e Amortizações","especieDocumento":"","dataReferencia":"26/02/2021","dataEntrega":"26/02/2021 17:34","status":"AC","descricaoStatus":"Ativo com visualização","analisado":"N","situacaoDocumento":"A","assuntos":null,"altaPrioridade":false,"formatoDataReferencia":"3","versao":1,"modalidade":"AP","descricaoModalidade":"Apresentação","nomePregao":"FII BB PRGII","informacoesAdicionais":"FII BB PRGII;","arquivoEstruturado":"","formatoEstruturaDocumento":null,"nomeAdministrador":null,"cnpjAdministrador":null,"cnpjFundo":null,"idTemplate":0,"idSelectNotificacaoConvenio":null,"idSelectItemConvenio":0,"indicadorFundoAtivoB3":false,"idEntidadeGerenciadora":null,"ofertaPublica":null,"numeroEmissao":null,"tipoPedido":null,"dda":null},{"id":142827,"descricaoFundo":"BB PROGRESSIVO II FUNDO DE INVESTIMENTO IMOBILIÁRIO - FII","categoriaDocumento":"Aviso aos Cotistas - Estruturado","tipoDocumento":"Rendimentos e Amortizações","especieDocumento":"","dataReferencia":"29/01/2021","dataEntrega":"29/01/2021 17:40","status":"AC","descricaoStatus":"Ativo com visualização","analisado":"N","situacaoDocumento":"A","assuntos":null,"altaPrioridade":false,"formatoDataReferencia":"3","versao":1,"modalidade":"AP","descricaoModalidade":"Apresentação","nomePregao":"FII BB PRGII","informacoesAdicionais":"FII BB PRGII;","arquivoEstruturado":"","formatoEstruturaDocumento":null,"nomeAdministrador":null,"cnpjAdministrador":null,"cnpjFundo":null,"idTemplate":0,"idSelectNotificacaoConvenio":null,"idSelectItemConvenio":0,"indicadorFundoAtivoB3":false,"idEntidadeGerenciadora":null,"ofertaPublica":null,"numeroEmissao":null,"tipoPedido":null,"dda":null}]}`
	resHTML = `<html>
<head>
<META http-equiv="Content-Type" content="text/html; charset=UTF-8">
<title>Informa&ccedil;&otilde;es sobre Pagamento de Proventos - FUNDOS</title>
<script type="text/javascript" src="/fnet/ruxitagentjs_ICA2QSVfhjqrux_10211210318124316.js" data-dtconfig="app=9b3c635ea455a875|rcdec=1209600000|featureHash=ICA2QSVfhjqrux|msl=153600|srsr=25000|vcv=2|rdnt=1|uxrgce=1|srcss=1|bp=3|srmcrv=10|cuc=qf4s7dpn|mel=100000|dpvc=1|md=mdcc1=a//*[@id^e^dquser-logoff-desktop^dq]/p/text(),mdcc2=a//*[@id^e^dqdivUsuario^dq],mdcc3=a#user-logoff-desktop ^rb p|lastModification=1617986327668|dtVersion=10211210318124316|srmcrl=1|tp=500,50,0,1|uxdcw=1500|vs=2|agentUri=/fnet/ruxitagentjs_ICA2QSVfhjqrux_10211210318124316.js|reportUrl=/fnet/rb_8370fec7-c82e-413f-a2c6-777046ed9811|rid=RID_-315536353|rpid=-2124372837|domain=bmfbovespa.com.br"></script><style type="text/css">
          table
          {
          border-collapse: collapse;
          }
          table, td, th
          {
          border: 1px solid black;
          font-size:12;
          }

          table.no_border
          {
          border-style: none;
          }

          tr.border_double td {
          border: 0px;
          border-bottom:2pt solid black;
          border-top:2pt solid black;
          }

          tr.no_border td {
          border: 0px;
          }


          body
          {
          margin:10px;
          font-family:"Times New Roman", Times, serif;
          }

          .titulo-tabela
          {
          display:block;
          font-weight:bold;
          text-align: center;
          font-size:14;
          }

          .titulo-dado
          {
          margin:5px;
          display:block;
          font-weight:bold;
          font-size:12;
          }

          .dado-cabecalho
          {
          margin:5px;
          display:block;
          font-size:12;
          }

          .dado-valores
          {
          margin:5px;
          display:block;
          font-size:12;
          text-align: center;
          }
        </style>
</head>
<body>
<a href="javascript:window.print()">Imprimir</a>
<h1 align="center">Informa&ccedil;&otilde;es sobre Pagamento de Proventos</h1>
<table border="1" width="95%" align="center">
<tr>
<td width="20%"><span class="titulo-dado">Nome do Fundo: </span></td><td width="40%"><span class="dado-cabecalho">BB PROGRESSIVO II FUNDO DE INVESTIMENTO IMOBILI&Aacute;RIO - FII</span></td><td width="20%"><span class="titulo-dado">CNPJ do Fundo: </span></td><td width="20%"><span class="dado-cabecalho">14.410.722/0001-29</span></td>
</tr>
<tr>
<td><span class="titulo-dado">Nome do Administrador: </span></td><td><span class="dado-cabecalho">VOTORANTIM ASSET MANAGEMENT DTVM LTDA.</span></td><td><span class="titulo-dado">CNPJ do Administrador: </span></td><td><span class="dado-cabecalho">03.384.738/0001-98</span></td>
</tr>
<tr>
<td><span class="titulo-dado">Respons&aacute;vel pela Informa&ccedil;&atilde;o: </span></td><td><span class="dado-cabecalho">XXX FII TEST XXX</span></td><td><span class="titulo-dado">Telefone Contato: </span></td><td><span class="dado-cabecalho">(11) 5171-5038</span></td>
</tr>
<tr>
<td><span class="titulo-dado">C&oacute;digo ISIN da cota: </span></td><td><span class="dado-cabecalho">BRBBPOCTF003</span></td><td><span class="titulo-dado">C&oacute;digo de negocia&ccedil;&atilde;o da cota: </span></td><td><span class="dado-cabecalho">BBPO11</span></td>
</tr>
</table>
<p></p>
<table cellpading="5" cellspacing="5" width="95%" align="center">
<tr>
<td width="60%"></td><td width="20%" valign="top" align="center"><b>Rendimento</b></td><td width="20%" align="center"><b>Amortiza&ccedil;&atilde;o</b></td>
</tr>
<tr>
<td>
          Ato societ&aacute;rio de aprova&ccedil;&atilde;o (se houver)
        </td><td><span class="dado-valores"></span></td><td><span class="dado-valores"></span></td>
</tr>
<tr>
<td>
          Data da informa&ccedil;&atilde;o
        </td><td><span class="dado-valores">31/03/2021</span></td><td></td>
</tr>
<tr>
<td>
          Data-base (&uacute;ltimo dia de negocia&ccedil;&atilde;o &ldquo;com&rdquo; direito ao provento)
        </td><td><span class="dado-valores">31/03/2021</span></td><td></td>
</tr>
<tr>
<td>
          Data do pagamento
        </td><td><span class="dado-valores">15/04/2021</span></td><td></td>
</tr>
<tr>
<td>
          Valor do provento por cota (R$)
        </td><td><span class="dado-valores">1,0823299</span></td><td></td>
</tr>
<tr>
<td>
          Per&iacute;odo de refer&ecirc;ncia
        </td><td><span class="dado-valores">mar&ccedil;o</span></td><td><span class="dado-valores"></span></td>
</tr>
<tr>
<td>
          Ano
        </td><td><span class="dado-valores">2021</span></td><td><span class="dado-valores"></span></td>
</tr>
<tr>
<td>
          Rendimento isento de IR*
        </td><td><span class="dado-valores">Sim</span></td><td></td>
</tr>
</table>
<p></p>
<table class="no_border" width="95%" align="center">
<tr class="no_border">
<td class="no_border"><span class="dado-valores">*A Administradora declara que o Fundo de Investimento Imobili&aacute;rio se enquadra no inciso III do art. 3&ordm; da Lei 11.033/2004, alterada pelo artigo 125 da Lei 11.196/2005. Em decorr&ecirc;ncia, fica isento do imposto de renda o cotista pessoa f&iacute;sica, desde que respeitado o disposto nos incisos I e II do par&aacute;grafo &uacute;nico do art. 3&ordm; da Lei 11.033/2004.</span></td>
</tr>
</table>
</body>
</html>
`
	resList = `"QUYgSU5WRVNUIENSSSBGRE8uIElOVi4gSU1PQiDigJMgUkVDRULDjVZFSVMgSU1PQi47RklJIEFGSEkgQ1JJOztBRkhJOwpBRiBJTlZFU1QgRkRPIElOVi4gSU1PQi4gLSBSRUNFQsONVkVJUyBJTU9CLjtGSUkgQUZJTlZDUjs7QUZDUjsKQUxJQU5aQSBGT0YgRlVORE8gREUgSU5WRVNUSU1FTlRPIElNT0JJTEnDgVJJTztGSUkgQUxJQU5aRkY7O0FGT0Y7CkFMSUFOWkEgTVVMVElPRkZJQ0VTICDigJMgRkRPLiBJTlYuIElNT0I7RklJIE1VTFQgT0YxOztNVE9GOwpBTElBTlpBIFRSVVNUIFJFTkRBIElNT0JJTElBUklBIEZETyBJTlYgSU1PQjtGSUkgQUxJQU5aQTs7QUxaUjsKQVEzIFJFTkRBIEZETyBJTlYgSU1PQiAtIEZJSTtGSUkgQVFVSUwgUkQ7TUI7QVJGSTsKw4FRVUlMTEEgRkRPIElOViBJTU9CIC0gRklJO0ZJSSBBUVVJTExBOztBUUxMOwpBVVRPTk9NWSBFRElGw41DSU9TIENPUlBPUkFUSVZPUyBGVU5ELiBJTlYuIElNT0IuO0ZJSSBBVVRPTk9NWTs7QUlFQzsKQkFORVNURVMgUkVDRULDjVZFSVMgSU1PQklMScOBUklPUyBGRE8gSU5WIElNT0IgIEZJSTtGSUkgQkVFUyBDUkk7O0JDUkk7CkJBTlJJU1VMIE5PVkFTIEZST05URUlSQVMgRkRPIElOViBJTU9CIC0gRklJO0ZJSSBCQU5SSVNVTDs7Qk5GUzsKQkFSWkVMIEZVTkRPIERFIElOVkVTVElNRU5UTyBJTU9CSUxJw4FSSU87RklJIEJBUlpFTDs7QlpFTDsKQkIgRkRPIElOViBJTU9CIFBST0dSRVNTSVZPO0ZJSSBCQiBQUk9HUjtNQjtCQkZJOwpCQiBGVU5ETyBERSBGVU5ET1MgLSBGRE8uIElOVi4gSU1PQi47RklJIEJCIEZPRjs7QkJGTzsK"`
	// resDetails = `{"detailFund":{"acronym":"AFHI","tradingName":"FII AFHI CRI","tradingCode":"AFHI11 ","tradingCodeOthers":"","cnpj":"36642293000158","classification":"Financeiro e Outros/Fundos/Fundos Imobiliários","webSite":"www.btgpactual.com/asset-management/administracao-fiduciaria","fundAddress":"N/I PRAIA DO BOTAFOGO 501 - 5 ANDAR, PARTE - CEP: 22250040 CIDADE: RIO DE JANEIRO UF: RJ","fundPhoneNumberDDD":"11","fundPhoneNumber":"33832513","fundPhoneNumberFax":"0","positionManager":"DIRETOR RESPONSAVEL","managerName":"ALLAN HADID","companyAddress":"N/I PRAIA DO BOTAFOGO 501 - 5 ANDAR, PARTE - CEP: 22250040 CIDADE: RIO DE JANEIRO UF: RJ","companyPhoneNumberDDD":"11","companyPhoneNumber":"33832513","companyPhoneNumberFax":"0","companyEmail":"ol-reguladores@btgpactual.com","companyName":"AF INVEST CRI FDO. INV. IMOB – RECEBÍVEIS IMOB.   ","quotaCount":"1787671","quotaDateApproved":"26/01/2021","codes":["AFHI11"],"codesOther":null,"segment":null},"shareHolder":{"shareHolderName":"BTG PACTUAL SERVIÇOS FINANC. S.A. DTVM","shareHolderAddress":"PRAIA DE BOTAFOGO 501 - 5.ANDAR               - CEP: 22250040 CIDADE: RIO DE JANEIRO UF: RJ","shareHolderPhoneNumberDDD":"21","shareHolderPhoneNumber":"32629600","shareHolderFaxNumber":"32628600","shareHolderEmail":"-"}}`
)

func newTestServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/pesquisarGerenciadorDocumentosDados", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resJSON))
	})

	mux.HandleFunc("/exibirDocumento", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(resHTML))
	})

	mux.HandleFunc("/fundsProxy/fundsCall/GetListFundDownload/eyJ0eXBlRnVuZCI6NywicGFnZU51bWJlciI6MSwicGFnZVNpemUiOjIwfQ==", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(resList))
	})

	mux.HandleFunc("/fundsProxy/fundsCall/GetDetailFundSIG", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(resList))
	})

	return httptest.NewServer(mux)
}

func TestFetchFII(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	if err := FetchFII(ts.URL); err != nil {
		t.Fatal(err)
	}
}

func TestMe(t *testing.T) {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get("https://sistemaswebb3-listados.b3.com.br/fundsProxy/fundsCall/GetListFundDownload/eyJ0eXBlRnVuZCI6NywicGFnZU51bWJlciI6MSwicGFnZVNpemUiOjIwfQ==")
	if err != nil {
		t.Error(err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	s, _ := strconv.Unquote(string(body))
	txt, err := base64.StdEncoding.DecodeString(s)

	log.Println("Body", string(txt), err)
}

func TestFetchFIIList(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	list, err := FetchFIIList(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	var whole string
	for i := range list {
		whole += list[i]
	}
	got := checksum(whole)
	want := [16]byte{0xd0, 0xc7, 0x34, 0x6b, 0x53, 0xe1, 0xa1, 0xbb, 0x57, 0x32, 0xf5, 0x9d, 0x55, 0xc5, 0x41, 0x75}
	if got != want {
		t.Fatalf("checksum expected %d, got %d", want, got)
	}
}

func checksum(b string) [16]byte {
	return md5.Sum([]byte(b))
}

func TestChecksum(t *testing.T) {
	b := [16]byte{0x10, 0x24, 0x8, 0xe4, 0x2b, 0x99, 0x27, 0x64, 0xc0, 0x8a, 0x7a, 0x2b, 0x3e, 0xa2, 0x9a, 0x4c}
	t.Log(checksum("any word") == b)
	t.Log(checksum("something longer"))
}

func TestFIIDetails(t *testing.T) {
	// ts := newTestServer()
	// defer ts.Close()

	ret, err := FetchFIIDetails("https://sistemaswebb3-listados.b3.com.br", "BBPO")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(prettyPrint(ret))
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func TestFII(t *testing.T) {
	const baseURL = "https://sistemaswebb3-listados.b3.com.br"

	list, err := FetchFIIList(baseURL)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(list)
	for i := range list {
		fii, err := FetchFIIDetails(baseURL, list[i])
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("- %s: %s", list[i], fii.DetailFund.Cnpj)
		if i > 10 {
			break
		}
	}

}
