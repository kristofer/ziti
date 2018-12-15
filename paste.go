package ziti

func (e *editor) setMark() {
	e.cb.mark.r, e.cb.mark.c = e.cb.point.r, e.cb.point.c
	e.cb.mark.ro, e.cb.mark.co = e.cb.point.ro, e.cb.point.co
	e.cb.markSet = true
	e.editorSetStatusMessage("Mark Set (%d,%d)", e.cb.mark.r+e.cb.mark.ro, e.cb.mark.c+e.cb.mark.co)
}
func (e *editor) noMark() bool {
	e.cb.markSet = false
	return e.cb.markSet
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
	sr, sc, er, ec, reverse := swapCursorsMaybe(e.cb.mark.r+e.cb.mark.ro, e.cb.mark.c+e.cb.mark.co, e.cb.point.ro+e.cb.point.r, e.cb.point.co+e.cb.point.c)
	for i := sr; i <= er; i++ {
		if sr == er {
			l := e.cb.rows[i]
			for j := sc; j < ec; j++ {
				e.pastebuffer = e.pastebuffer + string(l.runes[j])
			}
		} else if i == sr {
			l := e.cb.rows[i]
			for j := sc; j < l.size; j++ {
				e.pastebuffer = e.pastebuffer + string(l.runes[j])
			}
			e.pastebuffer = e.pastebuffer + "\n"
		} else if i == er {
			if i < e.cb.numrows {
				l := e.cb.rows[i]
				for j := 0; j < ec; j++ {
					e.pastebuffer = e.pastebuffer + string(l.runes[j])
				}
			}
		} else {
			l := e.cb.rows[i]
			for j := 0; j < l.size; j++ {
				e.pastebuffer = e.pastebuffer + string(l.runes[j])
			}
			e.pastebuffer = e.pastebuffer + "\n"
		}
	}
	// if del == true, remove all the chars
	if del == true {
		c2Remove := len(e.pastebuffer)
		if reverse { // move cursor to delete the right runes
			e.cb.point.r, e.cb.point.c = e.cb.mark.r, e.cb.mark.c
			e.cb.point.ro, e.cb.point.co = e.cb.mark.ro, e.cb.mark.co
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
