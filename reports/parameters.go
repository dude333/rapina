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
	// Reports is a map with the reports and reports items to be printed:
	// - ExtraRatios: enables some extra financial ratios on report
	// - ShowShares: shows the number of shares and free float on report
	// - Sector: creates a sheet with the sector report
	Reports map[string]bool
}
