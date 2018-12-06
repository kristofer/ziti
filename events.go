package ziti

import termbox "github.com/nsf/termbox-go"

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
	case CtrlK:
		/* We ignore ctrl-c, it can't be so simple to lose the changes to the edited file. */
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
	default:

	}
	//e.quitTimes = 2
}
