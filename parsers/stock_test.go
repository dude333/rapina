package parsers

import (
	"reflect"
	"strings"
	"testing"
)

func Test_parseB3(t *testing.T) {

	const file = `012021010412NSLU11      010FII LOURDES CI  ER       R$  000000002840000000000284000000000027700000000002809000000000281900000000028029000000002819000168000000000000001381000000000038793560000000000000009999123100000010000000000000BRNSLUCTF008272
012021010412NVHO11      010FII NOVOHORICI  ER       R$  000000000154000000000015900000000001535000000000153700000000015400000000001536000000000154000092000000000000006200000000000009533490000000000000009999123100000010000000000000BRNVHOCTF003186
012021010412ONEF11      010FII THE ONE CI           R$  000000001478800000000148000000000014717000000001478900000000147360000000014735000000001478700035000000000000002546000000000037652878000000000000009999123100000010000000000000BRONEFCTF003200`

	want := []stockQuote{
		{Stock: "NSLU11", Date: "2021-01-04", Open: 284, High: 284, Low: 277, Close: 281.9, Volume: 387935.6},
		{Stock: "NVHO11", Date: "2021-01-04", Open: 15.4, High: 15.9, Low: 15.35, Close: 15.4, Volume: 95334.9},
		{Stock: "ONEF11", Date: "2021-01-04", Open: 147.88, High: 148, Low: 147.17, Close: 147.36, Volume: 376528.78},
	}

	for i, line := range strings.Split(file, "\n") {
		got, err := parseB3Quote(line)
		if err != nil {
			t.Errorf("parseB3() error = %v", err)
			return
		}

		if err == nil && !reflect.DeepEqual(got, &want[i]) {
			t.Errorf("parseB3() got %+v, want %+v", got, &want[i])
		}

	}

}

func Test_parseB3Code(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name    string
		args    args
		want    *stockCode
		wantErr bool
	}{
		{
			name: "funds",
			args: args{
				`2021-05-13;ALMI11;ALMI;ALMI;CASH;EQUITY-CASH;FUNDS;;;2018-09-24;9999-12-31;;;;;BRALMICTF003;CICIRU;;;;;;1;BRL;;;;;;;;;;;;;;;;;250;1;2;;;;CI;FDO INV IMOB - FII TORRE ALMIRANTE;9999-12-31;FUNGIBLE;111177;`,
			},
			want:    &stockCode{TckrSymb: "ALMI11", SgmtNm: "CASH", SctyCtgyNm: "FUNDS", CrpnNm: "FDO INV IMOB - FII TORRE ALMIRANTE", SpcfctnCd: "CI", CorpGovnLvlNm: ""},
			wantErr: false,
		},
		{
			name: "unit",
			args: args{
				`2021-05-13;ALUP11;ALUP;ALUP;CASH;EQUITY-CASH;UNIT;;;2021-04-28;9999-12-31;;;;;BRALUPCDAM15;EMXXXR;;;;;;1;BRL;;;;;;;;;;;;;;;;;112;1;2;;;;UNT     N2;ALUPAR INVESTIMENTO S/A;9999-12-31;FUNGIBLE;136606616;NIVEL 2`,
			},
			want:    &stockCode{TckrSymb: "ALUP11", SgmtNm: "CASH", SctyCtgyNm: "UNIT", CrpnNm: "ALUPAR INVESTIMENTO S/A", SpcfctnCd: "UNT     N2", CorpGovnLvlNm: "NIVEL 2"},
			wantErr: false,
		},
		{
			name: "shares",
			args: args{
				`2021-05-13;ALPA3;ALPA;ALPA;CASH;EQUITY-CASH;SHARES;;;2020-02-17;9999-12-31;;;;;BRALPAACNOR0;ESVUFR;;;;;;1;BRL;;;;;;;;;;;;;;;;;229;1;2;;;;ON      N1;ALPARGATAS S.A.;9999-12-31;FUNGIBLE;302010689;NIVEL 1`,
			},
			want:    &stockCode{TckrSymb: "ALPA3", SgmtNm: "CASH", SctyCtgyNm: "SHARES", CrpnNm: "ALPARGATAS S.A.", SpcfctnCd: "ON      N1", CorpGovnLvlNm: "NIVEL 1"},
			wantErr: false,
		},
		{
			name: "should fail on odd lot",
			args: args{
				`2021-05-13;ANIM3F;ANIM;ANIM;ODD LOT;EQUITY-CASH;SHARES;;;2021-02-19;9999-12-31;;;;;BRANIMACNOR6;ESVUFR;;;;;;1;BRL;;;;;;;;;;;;;;;;;107;1;2;;;;ON      NM;ANIMA HOLDING S.A.;9999-12-31;FUNGIBLE;403868805;NOVO MERCADO`,
			},
			want:    &stockCode{},
			wantErr: true,
		},
		{
			name: "should fail on bdr",
			args: args{
				`2021-05-13;AMZO34;AMZO;AMZO;CASH;EQUITY-CASH;BDR;;;2020-11-09;9999-12-31;;;;;BRAMZOBDR002;EDSXPR;;;;;;1;BRL;;;;;;;;;;;;;;;;;102;1;2;;;;DRN;AMAZON.COM, INC;9999-12-31;FUNGIBLE;79059664651;`,
			},
			want:    &stockCode{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseB3Code(tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseB3Code() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseB3Code() = %#v, want %v", got, tt.want)
			}
		})
	}
}
