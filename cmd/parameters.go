package cmd

// Parms holds the input parameters
type Parms struct {
	// Company name to be processed
	Company string
	// OutputDir: path of the output xlsx
	OutputDir string
	// YamlFile: file with the companies' sectors
	YamlFile string
	// Reports is a map with the reports and reports items to be printed
	Reports map[string]bool
}
