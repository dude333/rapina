package reports

import (
	"encoding/json"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/pkg/errors"
)

// Used by style
const (
	NUMBER = iota + 1
	INDEX
	PERCENT
	EMPTY
	LEFT
	RIGHT
	DEFAULT
)

// Excel instance reachable data
type Excel struct {
	xlsx *excelize.File
}

//
// newExcel creates a new Excel instance
//
func newExcel() (e *Excel) {
	e = &Excel{}
	e.xlsx = excelize.NewFile()
	return
}

//
// saveAndCloseExcel saves to filename (need to set the directory as well)
//
func (e *Excel) saveAndCloseExcel(filename string) (err error) {
	// newFilename = time.Now().Format("02Jan06_150405.000") + ".xlsx" // DDMMMYY
	e.xlsx.DeleteSheet("Sheet1")
	e.xlsx.SetActiveSheet(1)
	err = e.xlsx.SaveAs(filename)
	if err != nil {
		return errors.Wrapf(err, "erro ao salvar planilha")
	}
	return
}

// Sheet struct
type Sheet struct {
	xlsx    *excelize.File
	name    string
	currRow int
}

func (e *Excel) newSheet(name string) (s *Sheet, err error) {
	s = &Sheet{}
	s.name = name
	s.xlsx = e.xlsx
	s.currRow = 1

	// Create a new sheet.
	// Avoid duplicated sheet
	if index := e.xlsx.GetSheetIndex(name); index > 0 {
		return nil, errors.Wrapf(err, "erro ao criar planilha %s", name)
	}

	e.xlsx.NewSheet(name)

	return
}

//
// printTitle prints the cols titles in Excel
//
func (s *Sheet) printTitle(cell string, title string) (err error) {
	// Print header
	s.xlsx.SetSheetRow(s.name, cell, &[]string{title})

	// Set styles
	style, err := s.xlsx.NewStyle(`{"number_format": 0,"font":{"bold":true},"alignment":{"horizontal":"center"},"border":[{"type":"bottom","color":"333333","style":3}]}`)
	if err == nil {
		s.xlsx.SetCellStyle(s.name, cell, cell, style)
	}

	return
}

//
// printRows prints cols in Excel
//
func (s *Sheet) printRows(startingCel string, slice *[]string, format int, bold bool) error {
	var err error
	var style int

	// Set styles
	json, err := jsonStyle(10, format, bold)
	if err != nil {
		return err
	}
	style, err = s.xlsx.NewStyle(string(json))
	if style > 0 && err == nil {
		col, row := cell2axis(startingCel)
		col += len(*slice)
		s.xlsx.SetCellStyle(s.name, startingCel, axis(col, row), style)
	}

	// Print row
	s.xlsx.SetSheetRow(s.name, startingCel, slice)

	return nil
}

//
// printValues prints cols in Excel
// Values >0 and <= 100 will be printed as %
//
func (s *Sheet) printValue(cell string, value float32, format int, bold bool) (err error) {

	s.xlsx.SetSheetRow(s.name, cell, &[]float32{value})

	// Set styles
	json, err := jsonStyle(10, format, bold)
	if err == nil {
		style, err := s.xlsx.NewStyle(string(json))
		if err == nil {
			s.xlsx.SetCellStyle(s.name, cell, cell, style)
		}
	}

	return nil
}

//
// printFormula
//
func (s *Sheet) printFormula(cell string, formula string, format int, bold bool) (err error) {

	s.xlsx.SetCellFormula(s.name, cell, formula)

	// Set styles
	json, err := jsonStyle(9, format, bold)
	if err == nil {
		style, err := s.xlsx.NewStyle(string(json))
		if err == nil {
			s.xlsx.SetCellStyle(s.name, cell, cell, style)
		}
	}

	return
}

//
// jsonStyle
//
func jsonStyle(size, format int, bold bool) ([]byte, error) {
	m := map[string]interface{}{
		"font": map[string]interface{}{"size": size, "bold": bold},
	}

	switch format {
	case PERCENT:
		m["custom_number_format"] = "0%;-0%;- "
	case INDEX:
		m["custom_number_format"] = "0.0;-0.0;-"
	case NUMBER:
		m["custom_number_format"] = "_-* #,##0,_-;_-* (#,##0,);_-* \"-\"_-;_-@_-"
	case RIGHT:
		m["alignment"] = map[string]interface{}{"horizontal": "right"}
	}

	j, err := json.Marshal(m)
	return j, err
}

//
// mergeCell
//
func (s *Sheet) mergeCell(a, b string) {
	s.xlsx.MergeCell(s.name, a, b)
}

//
// autoWidth adjust the cols width
//
func (s *Sheet) autoWidth() {
	const cols string = "ABCDEFGHIJKLMONPQRSTUVWXYZ"
	setColWidth := s.xlsx.SetColWidth
	setColWidth(s.name, "A", "A", 16)
	setColWidth(s.name, "B", "B", 48)

	// Get the space that separates the account numbers from the
	// vertical analysis numbers
	var spaced int
	for col := 2; col < len(cols); col++ {
		if len(s.xlsx.GetCellValue(s.name, axis(col, 1))) == 0 {
			spaced = col
			break
		}
	}

	setColWidth(s.name, "C", string(cols[spaced-1]), 9.5)                            // Account values
	setColWidth(s.name, string(cols[spaced]), string(cols[(spaced-3)+spaced]), 4.64) // Vertical Analysis values
}

func (s *Sheet) setColWidth(col int, width float64) {
	c := excelize.ToAlphaString(col)
	s.xlsx.SetColWidth(s.name, c, c, width)
}

type b struct {
	Borders []border `json:"border"`
}

type border struct {
	Type  string `json:"type"`
	Color string `json:"color"`
	Style int    `json:"style"`
}

func (s *Sheet) drawBorder(r1, c1, r2, c2, style int) {
	pos := []string{"top", "right", "bottom", "left"}
	ax := []struct {
		c int
		r int
	}{
		{c1, r1}, {c2, r1}, // top
		{c2, r1}, {c2, r2}, // right
		{c1, r2}, {c2, r2}, // bottom
		{c1, r1}, {c1, r2}, // left
	}
	for i, p := range pos {
		border := b{[]border{
			{p, "000000", style},
		}}

		json, err := json.Marshal(border)
		if err == nil {
			style, err := s.xlsx.NewStyle(string(json))
			if err == nil {
				j := i * 2
				s.xlsx.SetCellStyle(s.name, axis(ax[j].c, ax[j].r), axis(ax[j+1].c, ax[j+1].r), style)
			}
		}
	}

	// Corners [(top,right), (right,bottom), (bottom,left), (left,top)]
	for i := range []int{1, 2, 3, 0} {
		ii := i - 1
		if ii <= 0 {
			ii = len(pos) - 1
		}
		border := b{[]border{
			{pos[ii], "000000", style},
			{pos[i], "000000", style},
		}}

		json, err := json.Marshal(border)
		if err == nil {
			style, err := s.xlsx.NewStyle(string(json))
			if err == nil {
				j := i
				if i <= 2 {
					j = i * 2
				}
				s.xlsx.SetCellStyle(s.name, axis(ax[j].c, ax[j].r), axis(ax[j].c, ax[j].r), style)
			}
		}
	}

}

//
// axis transforms (2, 3) into "B3"
//
func axis(col, row int) string {
	return excelize.ToAlphaString(col) + strconv.Itoa(row)
}

//
// cell2axis only works from A1 to Z999
//
func cell2axis(cell string) (col, row int) {
	col = int(cell[0] - 'A')
	row, _ = strconv.Atoi(cell[1:])

	return
}
