package parsers

import (
	"fmt"

	"github.com/360EntSecGroup-Skylar/excelize"
)

// Excel instance reachable data
type Excel struct {
	xlsx *excelize.File
}

//
// openExcel opens an existing file
//
func openExcel(filename string) (e *Excel, err error) {
	e = &Excel{}
	e.xlsx, err = excelize.OpenFile(filename)
	return
}

func (e Excel) rows(sheet string) (lines [][]string, err error) {
	for _, row := range e.xlsx.GetRows(sheet) {
		for _, colCell := range row {
			fmt.Print(colCell, "|\t")
		}
		fmt.Println()
	}

	return
}
