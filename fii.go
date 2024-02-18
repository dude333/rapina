package rapina

// Dividend contains the stock 'Code', and the 'Date' for the stock dividend 'Val'.
type Dividend struct {
	Code        string
	Date        string
	PaymentDate string
	Val         float64
}

// Monthly contains the FII monthly report fields
type Monthly struct {
}

// FIIDetails details (ID field: DetailFund.CNPJ)
type FIIDetails struct {
	DetailFund struct {
		Acronym               string      `json:"acronym"`
		TradingName           string      `json:"tradingName"`
		TradingCode           string      `json:"tradingCode"`
		TradingCodeOthers     string      `json:"tradingCodeOthers"`
		CNPJ                  string      `json:"cnpj"`
		Classification        string      `json:"classification"`
		WebSite               string      `json:"webSite"`
		FundAddress           string      `json:"fundAddress"`
		FundPhoneNumberDDD    string      `json:"fundPhoneNumberDDD"`
		FundPhoneNumber       string      `json:"fundPhoneNumber"`
		FundPhoneNumberFax    string      `json:"fundPhoneNumberFax"`
		PositionManager       string      `json:"positionManager"`
		ManagerName           string      `json:"managerName"`
		CompanyAddress        string      `json:"companyAddress"`
		CompanyPhoneNumberDDD string      `json:"companyPhoneNumberDDD"`
		CompanyPhoneNumber    string      `json:"companyPhoneNumber"`
		CompanyPhoneNumberFax string      `json:"companyPhoneNumberFax"`
		CompanyEmail          string      `json:"companyEmail"`
		CompanyName           string      `json:"companyName"`
		QuotaCount            string      `json:"quotaCount"`
		QuotaDateApproved     string      `json:"quotaDateApproved"`
		Codes                 []string    `json:"codes"`
		CodesOther            interface{} `json:"codesOther"`
		Segment               interface{} `json:"segment"`
	} `json:"detailFund"`
	ShareHolder struct {
		ShareHolderName           string `json:"shareHolderName"`
		ShareHolderAddress        string `json:"shareHolderAddress"`
		ShareHolderPhoneNumberDDD string `json:"shareHolderPhoneNumberDDD"`
		ShareHolderPhoneNumber    string `json:"shareHolderPhoneNumber"`
		ShareHolderFaxNumber      string `json:"shareHolderFaxNumber"`
		ShareHolderEmail          string `json:"shareHolderEmail"`
	} `json:"shareHolder"`
}

// FIIStorage is the interface that contains the methods needed to parse, save and
// retrieve FII data to/from a storage.
type FIIStorage interface {
	Details(code string) (*FIIDetails, error)
	SaveDetails(stream []byte) error

	Dividends(code, monthYear string) (*[]Dividend, error)
	SaveDividend(dividend Dividend) error
}
