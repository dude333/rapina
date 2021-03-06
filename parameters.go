package rapina

// Parms holds the input parameters
type Parms struct {
	// Company name to be processed
	Company string
	// OutputDir: path of the output xlsx
	OutputDir string
	// YamlFile: file with the companies' sectors
	YamlFile string
	// ExtraRatios: enables some extra financial ratios on report
	ExtraRatios bool
	// ShowShares: shows the number of shares and free float on report
	ShowShares bool
	// OmitSector: omits the sector report
	OmitSector bool
}
