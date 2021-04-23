package parsers

/*
//
// FetchFIIs downloads the list of FIIs to get their code (e.g. 'HGLG'),
// then it uses this code to retrieve its details to get the CNPJ.
// Original baseURL: https://sistemaswebb3-listados.b3.com.br.
//
func FetchFIIList(baseURL string) ([]string, error) {
	listFundsURL := JoinURL(baseURL, `/fundsProxy/fundsCall/GetListFundDownload/eyJ0eXBlRnVuZCI6NywicGFnZU51bWJlciI6MSwicGFnZVNpemUiOjIwfQ==`)
	// fundsDetailsURL := `https://sistemaswebb3-listados.b3.com.br/fundsProxy/fundsCall/GetDetailFundSIG`

	tr := &http.Transport{
		DisableCompression: true,
		IdleConnTimeout:    30 * time.Second,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(listFundsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	unq, err := strconv.Unquote(string(body))
	if err != nil {
		return nil, err
	}
	txt, err := base64.StdEncoding.DecodeString(unq)
	if err != nil {
		return nil, err
	}

	var codes []string

	for _, line := range strings.Split(string(txt), "\n") {
		p := strings.Split(line, ";")
		if len(p) > 3 && len(p[3]) == 4 {
			codes = append(codes, p[3])
		}
	}

	return codes, nil
}


*/
