package ziti

import (
	termbox "github.com/nsf/termbox-go"
)

/* ========================= Editor events handling  ======================== */

/* Handle cursor position change because arrow keys were pressed. */
func (e *editor) editorMoveCursor(rch termbox.Key) {
	filerow := e.cb.point.ro + e.cb.point.r
	filecol := e.cb.point.co + e.cb.point.c
	var row *erow
	if filerow < e.cb.numrows {
		row = e.cb.rows[filerow]
	}
	switch rch {
	case ArrowLeft:
		if e.cb.point.c == 0 {
			if e.cb.point.co != 0 {
				e.cb.point.co--
			} else {
				if filerow > 0 {
					e.cb.point.r--
					e.cb.point.c = e.cb.rows[filerow-1].size
					if e.cb.point.c > e.screencols-1 {
						e.cb.point.co = e.cb.point.c - e.screencols + 1
						e.cb.point.c = e.screencols - 1
					}
				}
			}
		} else {
			e.cb.point.c--
		}
		break
	case ArrowRight:
		if row != nil && filecol < row.size {
			if e.cb.point.c == e.screencols-1 {
				e.cb.point.co++
			} else {
				e.cb.point.c++
			}
		} else if row != nil && filecol == row.size {
			e.cb.point.c = 0
			e.cb.point.co = 0
			if e.cb.point.r == e.screenrows-1 {
				e.cb.point.ro++
			} else {
				e.cb.point.r++
			}
		}
		break
	case ArrowUp:
		if e.cb.point.r == 0 {
			if e.cb.point.ro != 0 {
				e.cb.point.ro--
			}
		} else {
			e.cb.point.r--
		}
	case ArrowDown:
		if filerow < e.cb.numrows {
			if e.cb.point.r == e.screenrows-1 {
				e.cb.point.ro++
			} else {
				e.cb.point.r++
			}
		}
	default:
	}
	/* Fix cx if the currentLine line has not enough runes. */
	filerow = e.cb.point.ro + e.cb.point.r
	filecol = e.cb.point.co + e.cb.point.c
	if filerow < e.cb.numrows {
		row = e.cb.rows[filerow]
	}
	rowlen := 0
	if row != nil {
		rowlen = len(row.runes)
	}
	if filecol > rowlen {
		e.cb.point.c -= filecol - rowlen
		if e.cb.point.c < 0 {
			e.cb.point.co += e.cb.point.c
			e.cb.point.c = 0
		}
	}
	e.editorSetStatusMessage("point(%d,%d)[%d,%d]s(%d,%d)", e.cb.point.c, e.cb.point.r, e.cb.point.co, e.cb.point.ro, e.screencols, e.screenrows)
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
	case CtrlB:
		e.listBuffers()
	case CtrlC:
		e.cutCopy(false)
	case CtrlX:
		e.cutCopy(true)
	case CtrlV:
		e.paste()
	case CtrlY:
		e.loadHelp()
	case CtrlN:
		e.nextBuffer()
	case CtrlW:
		e.killBuffer()
	case CtrlO:
		e.loadFile()
	case CtrlA:
		e.movetoLineStart()
	case CtrlE:
		e.movetoLineEnd()
	case CtrlK:
		e.killtoLineEnd()
	case CtrlQ: /* Ctrl-q */
		/* Quit if the file was already saved. */
		if e.cb.dirty && e.quitTimes > 0 {
			e.editorSetStatusMessage("WARNING!!! File has unsaved changes. Press Ctrl-Q %d more times to quit.", e.quitTimes)
			e.quitTimes--
		}
		//fmt.Println("Done.")
		if e.quitTimes == 0 || !e.cb.dirty {
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
	case HomeKey:
		//e.editorSetStatusMessage("Home: Beginning of Buffer.")
		e.moveToBufferStart()
	case EndKey:
		e.editorSetStatusMessage("End: End of Buffer.")
		e.moveToBufferEnd()
	case PageUp, PageDown:
		if ev.Key == PageUp && e.cb.point.r != 0 {
			e.cb.point.r = 0
		} else {
			if ev.Key == PageDown && e.cb.point.r != e.screenrows-1 {
				e.cb.point.r = e.screenrows - 1
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
		e.editorRefreshScreen(true)
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
	if mr >= e.cb.numrows {
		mr = e.cb.numrows - 1
	}
	line := e.cb.rows[e.cb.point.r+e.cb.point.ro]
	if mc >= line.size {
		mc = line.size
	}
	if mc > 0 && mc < line.size {
		mc = e.mapRenderToRunes(line, mc)
		//log.Printf("mc %d\n", mc)
	}
	e.cb.point.c = mc
	e.cb.point.r = mr
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
