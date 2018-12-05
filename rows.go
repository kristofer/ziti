package ziti

/* ======================= Editor rows implementation ======================= */

/* Update the rendered version and the syntax highlight of a row. */
func (e *editor) editorUpdateRow(row *erow) {
	tabs := 0
	for j := 0; j < row.size; j++ {
		if row.runes[j] == Tab {
			tabs++
		}
	}
	row.render = make([]rune, len(row.runes)+tabs*4)
	row.rsize = len(row.render)
	//    /* Create a version of the row we can directly print on the screen,
	//      * respecting Tabs, substituting non prinTable characters with '?'. */
	idx := 0
	for j := 0; j < row.size; j++ {
		if row.runes[j] == Tab {
			row.render[idx] = ' '
			idx++
			for (idx+1)%4 != 0 {
				row.render[idx] = ' '
				idx++
			}
		} else {
			row.render[idx] = row.runes[j]
			idx++
		}
	}
	row.rsize = idx
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
	filerow := e.rowoff + e.cy
	filecol := e.coloff + e.cx
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
	if e.cx == e.screencols-1 {
		e.coloff++
	} else {
		e.cx++
	}
	e.dirty = true
}

/* Inserting a newline is slightly complex as we have to handle inserting a
 * newline in the middle of a line, splitting the line as needed. */
func (e *editor) editorInsertNewline() {
	filerow := e.rowoff + e.cy
	filecol := e.coloff + e.cx
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
	if e.cy == e.screenrows-1 {
		e.rowoff++
	} else {
		e.cy++
	}
	e.cx = 0
	e.coloff = 0
}

/* Delete the char at the currentLine cursor position. */
func (e *editor) editorDelChar() {
	filerow := e.rowoff + e.cy
	filecol := e.coloff + e.cx
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
		if e.cy == 0 {
			e.rowoff--
		} else {
			e.cy--
		}
		e.cx = filecol
		if e.cx >= e.screencols {
			shift := (e.screencols - e.cx) + 1
			e.cx -= shift
			e.coloff += shift
		}
	} else {
		e.editorRowDelChar(row, filecol-1)
		if e.cx == 0 && e.coloff != 0 {
			e.coloff--
		} else {
			e.cx--
		}
	}
	if row != nil {
		e.editorUpdateRow(row)
	}
	e.dirty = true
}

func (e *editor) movetoLineStart() {
	e.cx = 0
}
func (e *editor) movetoLineEnd() {
	e.cx = e.row[e.rowoff+e.cy].size
}
