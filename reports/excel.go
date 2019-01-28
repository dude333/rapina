package reports

import (
	"encoding/json"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/pkg/errors"
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
	xlsx *excelize.File
	name string
	row  int
	col  int
}

func (e *Excel) newSheet(name string) (s *Sheet, err error) {
	s = &Sheet{}
	s.name = name
	s.xlsx = e.xlsx

	// Create a new sheet.
	// Avoid duplicated sheet
	if index := e.xlsx.GetSheetIndex(name); index > 0 {
		return nil, errors.Wrapf(err, "erro ao criar planilha %s", name)
	}

	e.xlsx.NewSheet(name)

	return
}

func (s Sheet) printCell(row, col int, value interface{}, styleID int) {
	// Print value
	cell := axis(col, row)
	s.xlsx.SetCellValue(s.name, cell, value)

	// Format cell
	s.xlsx.SetCellStyle(s.name, cell, cell, styleID)
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
// print cols in Excel
//
func (s *Sheet) print(startingCel string, slice *[]string, format int, bold bool) error {
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

	// 012345678901234567890
	// AB2DE5GHI => 5-2+5 = 8
	// AB2DEF6HIJK => 6-2+6 = 10
	// AB2DEFG7IJKLM => 7-2+7 = 12
	setColWidth(s.name, "C", string(cols[spaced-1]), 9.5)                       // Account values
	setColWidth(s.name, string(cols[spaced]), string(cols[(spaced*2)-1]), 4.64) // Vertical Analysis values
}

func (s *Sheet) setColWidth(col int, width float64) {
	c := excelize.ToAlphaString(col)
	s.xlsx.SetColWidth(s.name, c, c, width)
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
