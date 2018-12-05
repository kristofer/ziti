package ziti

func (e *editor) setMark() {
	filerow := e.rowoff + e.cy
	filecol := e.coloff + e.cx
	e.marky = filerow
	e.markx = filecol
	e.editorSetStatusMessage("Mark Set (%d,%d)", filecol, filerow)
}
func (e *editor) cutCopy(del bool) {
	filerow := e.rowoff + e.cy
	filecol := e.coloff + e.cx
	e.pastebuffer = ""
	for i := e.marky; i <= filerow; i++ {
		if e.marky == filerow {
			l := e.row[i]
			for j := e.markx; j < filecol; j++ {
				e.pastebuffer = e.pastebuffer + string(l.runes[j])
			}
		} else if i == e.marky {
			l := e.row[i]
			for j := e.markx; j < l.size; j++ {
				e.pastebuffer = e.pastebuffer + string(l.runes[j])
			}
			e.pastebuffer = e.pastebuffer + "\n"
		} else if i == filerow {
			l := e.row[i]
			for j := 0; j < filecol; j++ {
				e.pastebuffer = e.pastebuffer + string(l.runes[j])
			}
		} else {
			l := e.row[i]
			for j := 0; j < l.size; j++ {
				e.pastebuffer = e.pastebuffer + string(l.runes[j])
			}
			e.pastebuffer = e.pastebuffer + "\n"
		}
	}
	// if del == true, remove all the chars
	if del == true {
		c2Remove := len(e.pastebuffer)
		for k := 0; k < c2Remove; k++ {
			e.editorDelChar()
		}
	}
}

func (e *editor) paste() {
	for _, rch := range e.pastebuffer {
		if rch == '\n' {
			e.editorInsertNewline()
		} else {
			e.editorInsertChar(rch)
		}
	}
}
