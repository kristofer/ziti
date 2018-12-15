package ziti

import (
	"fmt"
	"time"

	termbox "github.com/nsf/termbox-go"
)

/* ============================= Terminal update ============================ */

/* This function writes the whole screen using termbox-go */
func (e *editor) editorRefreshScreen() {
	termbox.Clear(termbox.ColorBlack, termbox.ColorDefault)
	// lastSelCol := -1
	// Draw the runes on the screen
	for y := 0; y < e.screenrows; y++ {
		filerow := e.cb.point.ro + y

		if filerow >= e.cb.numrows {
			drawline(y, e.fgcolor, e.bgcolor, "~")
		} else {
			r := e.cb.rows[filerow]
			len := r.rsize - e.cb.point.co
			if len > 0 {
				if len > e.screencols {
					len = e.screencols
				}
				for j := 0; j < len; j++ {
					runeToPrint := r.render[j]
					if r.render[j] == Tab {
						runeToPrint = ' '
					}
					if e.inSelectedRegion(j, filerow) {
						termbox.SetCell(j, y, runeToPrint, e.fgcolor|termbox.AttrUnderline, e.bgcolor)
						// lastSelCol = j
					} else {
						termbox.SetCell(j, y, runeToPrint, e.fgcolor, e.bgcolor)
					}
				}
			}
		}
	}
	/* Create a two rows for status. First row: */
	dirtyflag := ""
	if e.cb.dirty {
		dirtyflag = "(modified)"
	}
	status := fmt.Sprintf("-- %s - %d lines - %d/%d - %s",
		e.cb.filename, e.cb.numrows, e.cb.point.ro+e.cb.point.r+1, e.cb.point.co+e.cb.point.c+1, dirtyflag) //e.cb.dirty ? "(modified)" : "")
	slen := len(status)
	if slen > e.screencols {
		slen = e.screencols
	}
	for slen < e.screencols { // blank the rest of the line
		status = status + " "
		slen++
	}
	drawline(e.screenrows, e.fgcolor|termbox.AttrReverse, e.bgcolor|termbox.AttrReverse, status) //termbox.ColorWhite, termbox.ColorBlack, status)
	/* Second row: depends on e.statusmsg and the status message update time. */

	if len(e.statusmsg) > 0 && time.Since(e.statusmsgTime).Seconds() < 5 {
		drawline(e.screenrows+1, e.fgcolor, e.bgcolor, e.statusmsg)
	} else {
		drawline(e.screenrows+1, e.fgcolor, e.bgcolor, " ")
	}
	/* Put cursor at its currentLine position. Note that the horizontal position
	 * at which the cursor is displayed may be different compared to 'e.cb.point.c'
	 * because of Tabs. */
	cx := e.adjustCursor()
	filerow := e.cb.point.ro + e.cb.point.r
	for j := e.cb.point.co + e.cb.point.c; j < cx; j++ {
		if e.cb.markSet == true {
			termbox.SetCell(j, e.cb.point.r, e.cb.rows[filerow].render[j], e.fgcolor|termbox.AttrUnderline, e.bgcolor)
		}
	}

	termbox.SetCursor(cx, e.cb.point.r)
	termbox.Flush()
}

func (e *editor) inSelectedRegion(c, r int) bool {
	if e.cb.markSet == false {
		return false
	}
	sr, sc, er, ec, _ := swapCursorsMaybe(e.cb.mark.r+e.cb.mark.ro, e.cb.mark.c+e.cb.mark.co, e.cb.point.ro+e.cb.point.r, e.cb.point.co+e.cb.point.c)
	if r < sr || r > er {
		return false
	}
	if r == sr && r == er && c >= ec {
		return false
	}
	if r == sr && c >= sc {
		return true
	}
	if r == sr && c < sc {
		return false
	}
	if r > sr && r < er {
		return true
	}
	if r == er && c < ec {
		return true
	}
	return false
}

func (e *editor) adjustCursor() int {
	cx := 0
	filerow := e.cb.point.ro + e.cb.point.r
	var row *erow // := nil
	if filerow < e.cb.numrows {
		row = e.cb.rows[filerow]
	}
	if row != nil {
		for j := e.cb.point.co; j < (e.cb.point.c + e.cb.point.co); j++ {
			if j < row.size && row.runes[j] == Tab {
				cx = cx + 3 //7 - ((cx) % 8)
			}
			cx = cx + 1
		}
	}
	return cx
}
func drawline(y int, fg, bg termbox.Attribute, msg string) {
	x := 0
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

/* Set an editor status message for the second line of the status, at the
 * end of the screen. */
func (e *editor) editorSetStatusMessage(fm string, args ...interface{}) {
	e.statusmsg = fmt.Sprintf(fm, args...)
	e.statusmsgTime = time.Now()
	return
}
