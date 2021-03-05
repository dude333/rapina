package reports

import "database/sql"

// Parms holds the input parameters
type Parms struct {
	// DB database handle
	DB *sql.DB
	// Company name to be processed
	Company string
	// Filename: path and filename of the output xlsx
	Filename string
	// YamlFile: file with the companies' sectors
	YamlFile string
	// ExtraRatios: enables some extra financial ratios on report
	ExtraRatios bool
	// ShowShares: shows the number of shares and free float on report
	ShowShares bool
}
