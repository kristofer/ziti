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
			row.render[idx] = Tab //' '
			idx++
			for (idx)%tabWidth != 0 { // +1?
				row.render[idx] = Tab //' '
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
	if e.cb.readonly {
		e.readOnly()
		return
	}
	if at > e.cb.numrows {
		return
	}
	trow := &erow{}
	e.cb.rows = append(e.cb.rows, nil)
	if at != e.cb.numrows {
		for j := e.cb.numrows; j >= at+1; j-- {
			e.cb.rows[j] = e.cb.rows[j-1]
		}
	}
	trow.runes = []rune(s)
	trow.size = len(trow.runes)
	trow.render = []rune{}
	trow.rsize = 0
	trow.idx = at
	e.cb.rows[at] = trow
	e.editorUpdateRow(e.cb.rows[at])
	e.cb.numrows++
	e.cb.dirty = true
}

func (e *editor) editorDelRow(at int) {
	if e.cb.readonly {
		e.readOnly()
		return
	}
	if len(e.cb.rows) <= at {
		return
	}
	e.cb.rows = append(e.cb.rows[:at], e.cb.rows[at+1:]...)
	e.cb.numrows--
	e.cb.dirty = true
}

func (e *editor) editorRowsToString() string {
	ret := ""
	for j := 0; j < e.cb.numrows; j++ {
		ret = ret + string(e.cb.rows[j].runes) + "\n"
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
	if e.cb.readonly {
		e.readOnly()
		return
	}
	row.runes = insertRune(row.runes, at, rch)
	row.size = len(row.runes)
	e.editorUpdateRow(row)
	e.cb.dirty = true
}

/* Append the string 's' at the end of a row */
func (e *editor) editorRowAppendString(row *erow, s string) { //}, size_t len) {
	if e.cb.readonly {
		e.readOnly()
		return
	}
	row.runes = append(row.runes, []rune(s)...)
	row.size = len(row.runes)
	e.editorUpdateRow(row)
	e.cb.dirty = true
}

/* Delete the character at offset 'at' from the specified row. */
func (e *editor) editorRowDelChar(row *erow, at int) {
	if e.cb.readonly {
		e.readOnly()
		return
	}
	if row.size <= at {
		return
	}
	row.runes = append(row.runes[:at], row.runes[at+1:]...)
	row.size = len(row.runes)
	e.cb.dirty = true
}

/* Insert the specified char at the currentLine prompt position. */
func (e *editor) editorInsertChar(rch rune) {
	if e.cb.readonly {
		e.readOnly()
		return
	}
	filerow := e.cb.point.ro + e.cb.point.r
	filecol := e.cb.point.co + e.cb.point.c
	var row *erow
	if filerow < e.cb.numrows {
		row = e.cb.rows[filerow]
	}
	/* If the row where the cursor is currentLinely located does not exist in our
	 * logical representaion of the file, add enough empty rows as needed. */
	if row == nil {
		for e.cb.numrows <= filerow {
			e.editorInsertRow(e.cb.numrows, "")
		}
	}
	row = e.cb.rows[filerow]
	e.editorRowInsertChar(row, filecol, rch)
	if e.cb.point.c == e.screencols-1 {
		e.cb.point.co++
	} else {
		e.cb.point.c++
	}
	e.cb.dirty = true
}

/* Inserting a newline is slightly complex as we have to handle inserting a
 * newline in the middle of a line, splitting the line as needed. */
func (e *editor) editorInsertNewline() {
	filerow := e.cb.point.ro + e.cb.point.r
	filecol := e.cb.point.co + e.cb.point.c
	var row *erow
	if filerow < e.cb.numrows {
		row = e.cb.rows[filerow]
	}
	if row == nil {
		if filerow == e.cb.numrows {
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
		row = e.cb.rows[filerow]
		e.cb.rows[filerow].runes = row.runes[:filecol]
		row.size = len(e.cb.rows[filerow].runes)
		e.editorUpdateRow(row)
	}
	// fixcursor:
	e.fixcursor()
}

func (e *editor) fixcursor() {
	if e.cb.point.r == e.screenrows-1 {
		e.cb.point.ro++
	} else {
		e.cb.point.r++
	}
	e.cb.point.c = 0
	e.cb.point.co = 0
}

/* Delete the char at the currentLine cursor position. */
func (e *editor) editorDelChar() {
	if e.cb.readonly {
		e.readOnly()
		return
	}
	filerow := e.cb.point.ro + e.cb.point.r
	filecol := e.cb.point.co + e.cb.point.c
	var row *erow
	if filerow < e.cb.numrows {
		row = e.cb.rows[filerow]
	}
	if row == nil || (filecol == 0 && filerow == 0) {
		return
	}
	if filecol == 0 {
		/* Handle the case of column 0, we need to move the currentLine line
		 * on the right of the previous one. */
		filecol = e.cb.rows[filerow-1].size
		e.editorRowAppendString(e.cb.rows[filerow-1], string(row.runes))
		e.editorDelRow(filerow)
		row = nil
		if e.cb.point.r == 0 {
			e.cb.point.ro--
		} else {
			e.cb.point.r--
		}
		e.cb.point.c = filecol
		if e.cb.point.c >= e.screencols {
			shift := (e.screencols - e.cb.point.c) + 1
			e.cb.point.c -= shift
			e.cb.point.co += shift
		}
	} else {
		e.editorRowDelChar(row, filecol-1)
		if e.cb.point.c == 0 && e.cb.point.co != 0 {
			e.cb.point.co--
		} else {
			e.cb.point.c--
		}
	}
	if row != nil {
		e.editorUpdateRow(row)
	}
	e.cb.dirty = true
}

func (e *editor) movetoLineStart() {
	e.cb.point.c = 0
}
func (e *editor) movetoLineEnd() {
	e.cb.point.c = e.cb.rows[e.cb.point.ro+e.cb.point.r].size
}
func (e *editor) killtoLineEnd() {
	linesize := e.cb.rows[e.cb.point.ro+e.cb.point.r].size
	if linesize > 0 {
		e.movetoLineEnd()
		for k := 0; k < linesize; k++ {
			e.editorDelChar()
		}
	}
	if linesize == 0 {
		e.editorDelRow(e.cb.point.ro + e.cb.point.r)
	}
}
