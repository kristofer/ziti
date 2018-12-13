package ziti

import (
	termbox "github.com/nsf/termbox-go"
)

/* ========================= Editor events handling  ======================== */

/* Handle cursor position change because arrow keys were pressed. */
func (e *editor) editorMoveCursor(rch termbox.Key) {
	filerow := e.point.ro + e.point.r
	filecol := e.point.co + e.point.c
	var row *erow
	if filerow < e.numrows {
		row = e.row[filerow]
	}
	switch rch {
	case ArrowLeft:
		if e.point.c == 0 {
			if e.point.co != 0 {
				e.point.co--
			} else {
				if filerow > 0 {
					e.point.r--
					e.point.c = e.row[filerow-1].size
					if e.point.c > e.screencols-1 {
						e.point.co = e.point.c - e.screencols + 1
						e.point.c = e.screencols - 1
					}
				}
			}
		} else {
			e.point.c--
		}
		break
	case ArrowRight:
		if row != nil && filecol < row.size {
			if e.point.c == e.screencols-1 {
				e.point.co++
			} else {
				e.point.c++
			}
		} else if row != nil && filecol == row.size {
			e.point.c = 0
			e.point.co = 0
			if e.point.r == e.screenrows-1 {
				e.point.ro++
			} else {
				e.point.r++
			}
		}
		break
	case ArrowUp:
		if e.point.r == 0 {
			if e.point.ro != 0 {
				e.point.ro--
			}
		} else {
			e.point.r--
		}
	case ArrowDown:
		if filerow < e.numrows {
			if e.point.r == e.screenrows-1 {
				e.point.ro++
			} else {
				e.point.r++
			}
		}
	default:
	}
	/* Fix cx if the currentLine line has not enough runes. */
	filerow = e.point.ro + e.point.r
	filecol = e.point.co + e.point.c
	if filerow < e.numrows {
		row = e.row[filerow]
	}
	rowlen := 0
	if row != nil {
		rowlen = len(row.runes)
	}
	if filecol > rowlen {
		e.point.c -= filecol - rowlen
		if e.point.c < 0 {
			e.point.co += e.point.c
			e.point.c = 0
		}
	}
}

/* Process events arriving from the standard input, which is, the user
 * is typing stuff on the terminal. */
func (e *editor) editorProcessEvent(ev termbox.Event) {
	//log.Printf("editorProcessEvent %#v\n", ev)
	if ev.Ch != 0 {
		e.editorInsertChar(ev.Ch)
		return
	}
	char := int16(ev.Key)
	switch char {
	case Space:
		e.editorInsertChar(' ')
	case Tab:
		e.editorInsertChar('\t')
	case Enter: /* Enter */
		e.editorInsertNewline()
	case CtrlC:
		e.cutCopy(false)
	case CtrlX:
		e.cutCopy(true)
	case CtrlV:
		e.paste()
	case CtrlA:
		e.movetoLineStart()
	case CtrlE:
		e.movetoLineEnd()
	case CtrlK:
		e.killtoLineEnd()
	case CtrlQ: /* Ctrl-q */
		/* Quit if the file was already saved. */
		if e.dirty && e.quitTimes > 0 {
			e.editorSetStatusMessage("WARNING!!! File has unsaved changes. Press Ctrl-Q %d more times to quit.", e.quitTimes)
			e.quitTimes--
		}
		//fmt.Println("Done.")
		if e.quitTimes == 0 || !e.dirty {
			e.done = true
		}
	case CtrlS: /* Ctrl-s */
		//fmt.Println("Save.")
		e.editorSave()
	case CtrlF:
		e.editorFind()
	case Backspace, CtrlH:
		e.editorDelChar()
	default:
	}

	switch ev.Key {
	case KeyNull:
		e.setMark()
	case DelKey:
		e.editorDelChar()
	case PageUp, PageDown:
		if ev.Key == PageUp && e.point.r != 0 {
			e.point.r = 0
		} else {
			if ev.Key == PageDown && e.point.r != e.screenrows-1 {
				e.point.r = e.screenrows - 1
			}
		}
		times := e.screenrows
		for times > 0 {
			if ev.Key == PageUp {
				e.editorMoveCursor(ArrowUp)
			} else {
				e.editorMoveCursor(ArrowDown)
			}
			times--
		}
	case ArrowUp, ArrowDown, ArrowLeft, ArrowRight:
		e.editorMoveCursor(ev.Key)
	case CtrlL: /* ctrl+l, clear screen */
		e.editorRefreshScreen()
	case Esc:
		/* Nothing to do for Esc in this mode. */
	case termbox.KeyCtrlG:
		e.editorSetStatusMessage("Mark cleared.")
		e.noMark()
	default:

	}
	//e.quitTimes = 2

	switch ev.Type {
	case termbox.EventMouse:
		//termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		e.editorSetStatusMessage("Mouse: c %d, r %d ", ev.MouseX+1, ev.MouseY+1)
		e.setPointForMouse(ev.MouseX, ev.MouseY)
		return
	case termbox.EventResize:
		e.resize()
		return
	default:
	}
}

func (e *editor) setPointForMouse(mc, mr int) {
	if mr > e.screenrows {
		mr = e.screenrows
	}
	if mr >= e.numrows {
		mr = e.numrows - 1
	}
	line := e.row[e.point.r+e.point.ro]
	if mc >= line.size {
		mc = line.size
	}
	if mc > 0 && mc < line.size {
		mc = e.mapRenderToRunes(line, mc)
		//log.Printf("mc %d\n", mc)
	}
	e.point.c = mc
	e.point.r = mr
}

// mapRenderToRunes should map screencolumn to runes column through the render line
func (e *editor) mapRenderToRunes(row *erow, col int) int {
	nc := 0
	row.rsize = len(row.render)
	hits := make([]int, row.rsize)
	idx := 0
	for j := 0; j < row.size; j++ {
		if row.runes[j] == Tab {
			hits[idx] = j
			idx++
			for (idx)%tabWidth != 0 { // +1?
				hits[idx] = j
				idx++
			}
		} else {
			hits[idx] = j
			idx++
		}
	}
	nc = hits[col]
	return nc
}
