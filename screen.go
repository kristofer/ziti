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
	// Draw the runes on the screen
	for y := 0; y < e.screenrows; y++ {
		filerow := e.rowoff + y

		if filerow >= e.numrows {
			drawline(y, e.fgcolor, e.bgcolor, "~")
		} else {
			r := e.row[filerow]
			len := r.rsize - e.coloff
			if len > 0 {
				if len > e.screencols {
					len = e.screencols
				}

				for j := 0; j < len; j++ {

					termbox.SetCell(j, y, r.render[j], e.fgcolor, e.bgcolor)
				}
			}
		}
	}
	/* Create a two rows for status. First row: */
	dirtyflag := ""
	if e.dirty {
		dirtyflag = "(modified)"
	}
	status := fmt.Sprintf("-- %s - %d lines - %d/%d - %s",
		e.filename, e.numrows, e.rowoff+e.cy+1, e.coloff+e.cx+1, dirtyflag) //e.dirty ? "(modified)" : "")
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
	 * at which the cursor is displayed may be different compared to 'e.cx'
	 * because of Tabs. */
	j := 0
	cx := 0
	filerow := e.rowoff + e.cy
	var row *erow // := nil
	if filerow < e.numrows {
		row = e.row[filerow]
	}
	if row != nil {
		for j = e.coloff; j < (e.cx + e.coloff); j++ {
			if j < row.size && row.runes[j] == Tab {
				cx += 7 - ((cx) % 8)
			}
			cx++
		}
	}
	termbox.SetCursor(cx, e.cy)
	termbox.Flush()
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
