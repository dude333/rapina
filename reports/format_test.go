package reports

import (
	"testing"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func TestFormat(t *testing.T) {
	var f [2]*formatStyle
	var style [2]int
	xlsx := excelize.NewFile()

	f[0] = newFormat(NUMBER, LEFT, false)
	f[1] = newFormat(INDEX, RIGHT, false)

	f[0].NumFmt = 10
	f[1].Lang = "en-US"

	for i := range f {
		style[i] = f[i].newStyle(xlsx)
		if style[i] == 0 {
			t.Error("Expecting style > 0, received 0")
		}
	}

	var f2 [2]*formatStyle
	var style2 [2]int

	f2[0] = newFormat(NUMBER, LEFT, false)
	f2[1] = newFormat(INDEX, RIGHT, false)

	f2[0].NumFmt = 10
	f2[1].Lang = "en-US"

	for i := range f {
		style2[i] = f[i].newStyle(xlsx)
		if style2[i] != style[i] {
			t.Errorf("Expecting style == %d, received %d", style[i], style2[i])
		}
	}
}
