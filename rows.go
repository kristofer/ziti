package ziti

/* ======================= Editor rows implementation ======================= */

/* Update the rendered version and the syntax highlight of a row. */
func (e *editor) editorUpdateRow(row *erow) {
	tabs := e.countTabs(row, row.size-1)
	row.render = make([]rune, len(row.runes)+tabs*tabWidth)
	row.rsize = len(row.render)
	/* Create a version of the row we can directly print on the screen */
	idx := 0
	for j := 0; j < row.size; j++ {
		if row.runes[j] == Tab {
			row.render[idx] = ' '
			idx++
			for (idx)%tabWidth != 0 { // +1?
				row.render[idx] = ' '
				idx++
			}
		} else {
			row.render[idx] = row.runes[j]
			idx++
		}
	}
	//row.rsize = idx
}

func (e *editor) countTabs(row *erow, toCol int) int {
	tabs := 0
	for j := 0; j <= toCol; j++ {
		if row.runes[j] == Tab {
			tabs++
		}
	}
	return tabs
}

func (e *editor) negativeOffsetFor(tabs int) int {
	return -1.0 * (tabs * tabWidth)
}

/* Insert a row at the specified position, shifting the other rows on the bottom
 * if required. */
func (e *editor) editorInsertRow(at int, s string) {
	if at > e.numrows {
		return
	}
	trow := &erow{}
	e.row = append(e.row, nil)
	if at != e.numrows {
		for j := e.numrows; j >= at+1; j-- {
			e.row[j] = e.row[j-1]
		}
	}
	trow.runes = []rune(s)
	trow.size = len(trow.runes)
	trow.render = []rune{}
	trow.rsize = 0
	trow.idx = at
	e.row[at] = trow
	e.editorUpdateRow(e.row[at])
	e.numrows++
	e.dirty = true
}

func (e *editor) editorDelRow(at int) {
	if len(e.row) <= at {
		return
	}
	e.row = append(e.row[:at], e.row[at+1:]...)
	e.numrows--
	e.dirty = true
}

func (e *editor) editorRowsToString() string {
	ret := ""
	for j := 0; j < e.numrows; j++ {
		ret = ret + string(e.row[j].runes) + "\n"
	}
	return ret
}

func insertRune(slice []rune, index int, value rune) []rune {
	// Grow the slice by one element.
	//slice = slice[0 : len(slice)+1]
	slice = append(slice, ' ')
	// Use copy to move the upper part of the slice out of the way and open a hole.
	copy(slice[index+1:], slice[index:])
	// Store the new value.
	slice[index] = value
	// Return the result.
	return slice
}

/* Insert a character at the specified position in a row, moving the remaining
 * runes on the right if needed. */
func (e *editor) editorRowInsertChar(row *erow, at int, rch rune) {
	row.runes = insertRune(row.runes, at, rch)
	row.size = len(row.runes)
	e.editorUpdateRow(row)
	e.dirty = true
}

/* Append the string 's' at the end of a row */
func (e *editor) editorRowAppendString(row *erow, s string) { //}, size_t len) {
	row.runes = append(row.runes, []rune(s)...)
	row.size = len(row.runes)
	e.editorUpdateRow(row)
	e.dirty = true
}

/* Delete the character at offset 'at' from the specified row. */
func (e *editor) editorRowDelChar(row *erow, at int) {
	if row.size <= at {
		return
	}
	row.runes = append(row.runes[:at], row.runes[at+1:]...)
	row.size = len(row.runes)
	e.dirty = true
}

/* Insert the specified char at the currentLine prompt position. */
func (e *editor) editorInsertChar(rch rune) {
	filerow := e.point.ro + e.point.r
	filecol := e.point.co + e.point.c
	var row *erow
	if filerow < e.numrows {
		row = e.row[filerow]
	}
	/* If the row where the cursor is currentLinely located does not exist in our
	 * logical representaion of the file, add enough empty rows as needed. */
	if row == nil {
		for e.numrows <= filerow {
			e.editorInsertRow(e.numrows, "")
		}
	}
	row = e.row[filerow]
	e.editorRowInsertChar(row, filecol, rch)
	if e.point.c == e.screencols-1 {
		e.point.co++
	} else {
		e.point.c++
	}
	e.dirty = true
}

/* Inserting a newline is slightly complex as we have to handle inserting a
 * newline in the middle of a line, splitting the line as needed. */
func (e *editor) editorInsertNewline() {
	filerow := e.point.ro + e.point.r
	filecol := e.point.co + e.point.c
	var row *erow
	if filerow < e.numrows {
		row = e.row[filerow]
	}
	if row == nil {
		if filerow == e.numrows {
			e.editorInsertRow(filerow, "")
			e.fixcursor()
		}
		return
	}
	/* If the cursor is over the currentLine line size, we want to conceptually
	 * think it's just over the last character. */
	if filecol >= row.size {
		filecol = row.size
	}
	if filecol == 0 {
		e.editorInsertRow(filerow, "")
	} else {
		/* We are in the middle of a line. Split it between two rows. */
		e.editorInsertRow(filerow+1, string(row.runes[filecol:]))
		row = e.row[filerow]
		e.row[filerow].runes = row.runes[:filecol]
		row.size = len(e.row[filerow].runes)
		e.editorUpdateRow(row)
	}
	// fixcursor:
	e.fixcursor()
}

func (e *editor) fixcursor() {
	if e.point.r == e.screenrows-1 {
		e.point.ro++
	} else {
		e.point.r++
	}
	e.point.c = 0
	e.point.co = 0
}

/* Delete the char at the currentLine cursor position. */
func (e *editor) editorDelChar() {
	filerow := e.point.ro + e.point.r
	filecol := e.point.co + e.point.c
	var row *erow
	if filerow < e.numrows {
		row = e.row[filerow]
	}
	if row == nil || (filecol == 0 && filerow == 0) {
		return
	}
	if filecol == 0 {
		/* Handle the case of column 0, we need to move the currentLine line
		 * on the right of the previous one. */
		filecol = e.row[filerow-1].size
		e.editorRowAppendString(e.row[filerow-1], string(row.runes))
		e.editorDelRow(filerow)
		row = nil
		if e.point.r == 0 {
			e.point.ro--
		} else {
			e.point.r--
		}
		e.point.c = filecol
		if e.point.c >= e.screencols {
			shift := (e.screencols - e.point.c) + 1
			e.point.c -= shift
			e.point.co += shift
		}
	} else {
		e.editorRowDelChar(row, filecol-1)
		if e.point.c == 0 && e.point.co != 0 {
			e.point.co--
		} else {
			e.point.c--
		}
	}
	if row != nil {
		e.editorUpdateRow(row)
	}
	e.dirty = true
}

func (e *editor) movetoLineStart() {
	e.point.c = 0
}
func (e *editor) movetoLineEnd() {
	e.point.c = e.row[e.point.ro+e.point.r].size
}
func (e *editor) killtoLineEnd() {
	linesize := e.row[e.point.ro+e.point.r].size
	if linesize > 0 {
		e.movetoLineEnd()
		for k := 0; k < linesize; k++ {
			e.editorDelChar()
		}
	}
	if linesize == 0 {
		e.editorDelRow(e.point.ro + e.point.r)
	}
}
