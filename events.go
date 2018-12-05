package ziti

import termbox "github.com/nsf/termbox-go"

/* ========================= Editor events handling  ======================== */

/* Handle cursor position change because arrow keys were pressed. */
func (e *editor) editorMoveCursor(rch termbox.Key) {
	filerow := e.rowoff + e.cy
	filecol := e.coloff + e.cx
	var row *erow
	if filerow < e.numrows {
		row = e.row[filerow]
	}
	switch rch {
	case ArrowLeft:
		if e.cx == 0 {
			if e.coloff != 0 {
				e.coloff--
			} else {
				if filerow > 0 {
					e.cy--
					e.cx = e.row[filerow-1].size
					if e.cx > e.screencols-1 {
						e.coloff = e.cx - e.screencols + 1
						e.cx = e.screencols - 1
					}
				}
			}
		} else {
			e.cx -= 1
		}
		break
	case ArrowRight:
		if row != nil && filecol < row.size {
			if e.cx == e.screencols-1 {
				e.coloff++
			} else {
				e.cx += 1
			}
		} else if row != nil && filecol == row.size {
			e.cx = 0
			e.coloff = 0
			if e.cy == e.screenrows-1 {
				e.rowoff++
			} else {
				e.cy += 1
			}
		}
		break
	case ArrowUp:
		if e.cy == 0 {
			if e.rowoff != 0 {
				e.rowoff--
			}
		} else {
			e.cy -= 1
		}
	case ArrowDown:
		if filerow < e.numrows {
			if e.cy == e.screenrows-1 {
				e.rowoff++
			} else {
				e.cy += 1
			}
		}
	default:
	}
	/* Fix cx if the currentLine line has not enough runes. */
	filerow = e.rowoff + e.cy
	filecol = e.coloff + e.cx
	if filerow < e.numrows {
		row = e.row[filerow]
	}
	rowlen := 0
	if row != nil {
		rowlen = len(row.runes)
	}
	if filecol > rowlen {
		e.cx -= filecol - rowlen
		if e.cx < 0 {
			e.coloff += e.cx
			e.cx = 0
		}
	}
}

/* Process events arriving from the standard input, which is, the user
 * is typing stuff on the terminal. */
func (e *editor) editorProcessEvent(ev termbox.Event) {

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
	case CtrlC: /* Ctrl-c */
		e.cutCopy(false)
	case CtrlX: /* Ctrl-c */
		e.cutCopy(true)
	case CtrlV: /* Ctrl-c */
		e.paste()
	case CtrlA: /* Ctrl-c */
		/* We ignore ctrl-c, it can't be so simple to lose the changes to the edited file. */
		e.movetoLineStart()
	case CtrlE: /* Ctrl-c */
		/* We ignore ctrl-c, it can't be so simple to lose the changes to the edited file. */
		e.movetoLineEnd()
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
		if ev.Key == PageUp && e.cy != 0 {
			e.cy = 0
		} else {
			if ev.Key == PageDown && e.cy != e.screenrows-1 {
				e.cy = e.screenrows - 1
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
	default:

	}
	//e.quitTimes = 2
}
