package ziti

func (e *editor) setMark() {
	e.mark.r, e.mark.c = e.cy, e.cx
	e.mark.ro, e.mark.co = e.rowoff, e.coloff
	e.markSet = true
	e.editorSetStatusMessage("Mark Set (%d,%d)", e.mark.r+e.mark.ro, e.mark.c+e.mark.co)
}
func (e *editor) noMark() bool {
	e.markSet = false
	return e.markSet
}
func swapCursorsMaybe(mr, mc, cr, cc int) (sr, sc, er, ec int, f bool) {
	if mr == cr {
		if mc > cc {
			return cr, cc, mr, mc, true
		}
		return mr, mc, cr, cc, false
	}
	if mr > cr {
		return cr, cc, mr, mc, true
	}
	return mr, mc, cr, cc, false
}

func (e *editor) cutCopy(del bool) {
	if e.noMark() == true {
		return
	}
	e.pastebuffer = ""
	sr, sc, er, ec, reverse := swapCursorsMaybe(e.mark.r+e.mark.ro, e.mark.c+e.mark.co, e.rowoff+e.cy, e.coloff+e.cx)
	for i := sr; i <= er; i++ {
		if sr == er {
			l := e.row[i]
			for j := sc; j < ec; j++ {
				e.pastebuffer = e.pastebuffer + string(l.runes[j])
			}
		} else if i == sr {
			l := e.row[i]
			for j := sc; j < l.size; j++ {
				e.pastebuffer = e.pastebuffer + string(l.runes[j])
			}
			e.pastebuffer = e.pastebuffer + "\n"
		} else if i == er {
			l := e.row[i]
			for j := 0; j < ec; j++ {
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
		if reverse {
			e.cy, e.cx = e.mark.r, e.mark.c
			e.rowoff, e.coloff = e.mark.ro, e.mark.co
		}
		for k := 0; k < c2Remove; k++ {
			e.editorDelChar()
		}
	}
	e.noMark()
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
