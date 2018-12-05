package ziti

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"time"

	termbox "github.com/nsf/termbox-go"
)

/* Ziti -- A very simple editor in less than 1000 lines of Go code .
 *
 * -----------------------------------------------------------------------
 *
 * Copyright (c) 2018, Kristofer Younger <kryounger at gmail dot com>
 * (who ripped out all the highlighting stuff)
 * based on the work of https://github.com/antirez/kilo
 * which was marked
 * Copyright (C) 2016 Salvatore Sanfilippo <antirez at gmail dot com>
 *
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 *
 *  *  Redistributions of source code must retain the above copyright
 *     notice, this list of conditions and the following disclaimer.
 * met:
 *
 *  *  Redistributions in binary form must reproduce the above copyright
 *     notice, this list of conditions and the following disclaimer in the
 *     documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTabILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES LOSS OF USE,
 * DATA, OR PROFITS OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

const zitiVersion = "1.1"

type Ziti struct {
	ziti *editor
}

/* This structure represents a single line of the file we are editing. */
type erow struct {
	idx    int    /* Row index in the file, zero-based. */
	size   int    /* Size of the row, excluding the null term. */
	rsize  int    /* Size of the rendered row. */
	runes  []rune //string /* Row content. */
	render []rune /* Row content "rendered" for screen (for Tabs). */
}

type editor struct {
	events        chan termbox.Event
	cx            int
	cy            int /* Cursor x and y position in characters */
	rowoff        int /* Offset of row displayed. */
	markx         int /* set mark for copy */
	marky         int
	coloff        int /* Offset of column displayed. */
	screenrows    int /* Number of rows that we can show */
	screencols    int /* Number of cols that we can show */
	numrows       int /* Number of rows */
	TextSize      int
	quitTimes     int
	done          bool
	rawmode       bool    /* Is terminal raw mode enabled? */
	row           []*erow /* Rows */
	dirty         bool    /* File modified but not saved. */
	filename      string  /* Currently open filename */
	statusmsg     string
	statusmsgTime time.Time
	pastebuffer   string
}

func (e *editor) checkErr(er error) {
	if er != nil {
		e.editorSetStatusMessage("%s", er)
	}
}

// KEY constants
const (
	KeyNull   = 0  /* NULL ctrl-space set mark */
	CtrlA     = 1  /* Ctrl-a BOL */
	CtrlC     = 3  /* Ctrl-c  cop */
	CtrlE     = 5  /* Ctrl-e  EOL */
	CtrlD     = 4  /* Ctrl-d del forward? */
	CtrlF     = 6  /* Ctrl-f find */
	CtrlH     = 8  /* Ctrl-h del backward*/
	Tab       = 9  /* Tab */
	CtrlL     = 12 /* Ctrl+l redraw */
	Enter     = 13 /* Enter */
	CtrlQ     = 17 /* Ctrl-q quit*/
	CtrlS     = 19 /* Ctrl-s save*/
	CtrlU     = 21 /* Ctrl-u number of times??*/
	CtrlV     = 22 /* Ctrl-V paste */
	CtrlW     = 23
	CtrlX     = 24 /* Ctrl-X cut */
	CtrlY     = 25
	CtrlZ     = 26
	Esc       = 27  /* Escape */
	Space     = 32  /* Space */
	Backspace = 127 /* Backspace */
)

// Cursor movement keys
const (
	ArrowLeft  = termbox.KeyArrowLeft
	ArrowRight = termbox.KeyArrowRight
	ArrowUp    = termbox.KeyArrowUp
	ArrowDown  = termbox.KeyArrowDown
	DelKey     = termbox.KeyDelete
	HomeKey    = termbox.KeyHome
	EndKey     = termbox.KeyEnd
	PageUp     = termbox.KeyPgup
	PageDown   = termbox.KeyPgdn
)

// State contains the state of a terminal.
type State struct {
	termios syscall.Termios
}

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

/* Delete the char at the currentLine prompt position. */
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

func (e *editor) movetoLineStart() {
	e.cx = 0
}
func (e *editor) movetoLineEnd() {
	e.cx = e.row[e.rowoff+e.cy].size
}

/* =================== FILE OPS ==============================*/
/* Load the specified program in the editor memory and returns 0 on success
 * or 1 on error. */
func (e *editor) editorOpen(filename string) error {

	// open the file filename
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	e.filename = filename
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// does the line contain the newline?
		line := scanner.Text()
		e.editorInsertRow(e.numrows, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	e.dirty = false
	return nil
}

/* Save the currentLine file on disk. Return 0 on success, 1 on error. */
func (e *editor) editorSave() error {
	buf := e.editorRowsToString()
	err := ioutil.WriteFile(e.filename, []byte(buf), 0644)
	e.checkErr(err)
	if err == nil {
		e.editorSetStatusMessage("Saved %d bytes (%d runes).", len(buf), len([]rune(buf)))
		e.dirty = false
	}
	return err
}

/* ============================= Terminal update ============================ */

/* This function writes the whole screen using termbox-go */
func (e *editor) editorRefreshScreen() {
	termbox.Clear(termbox.ColorBlack, termbox.ColorDefault)
	for y := 0; y < e.screenrows; y++ {
		filerow := e.rowoff + y

		if filerow >= e.numrows {
			drawline(y, termbox.ColorBlack, termbox.ColorDefault, "~")
		} else {
			r := e.row[filerow]
			len := r.rsize - e.coloff
			if len > 0 {
				if len > e.screencols {
					len = e.screencols
				}

				for j := 0; j < len; j++ {

					termbox.SetCell(j, y, r.render[j], termbox.ColorBlack, termbox.ColorDefault)
				}
			}
		}
	}
	/* Create a two rows status. First row: */
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

	for slen < e.screencols {
		status = status + " "
		slen++
	}
	drawline(e.screenrows, termbox.ColorWhite, termbox.ColorBlack, status)
	/* Second row depends on e.statusmsg and the status message update time. */

	if len(e.statusmsg) > 0 && time.Since(e.statusmsgTime).Seconds() < 5 {
		drawline(e.screenrows+1, termbox.ColorBlack, termbox.ColorDefault, e.statusmsg)
	} else {
		drawline(e.screenrows+1, termbox.ColorBlack, termbox.ColorDefault, " ")
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

/* =============================== Find mode ================================ */

func (e *editor) runeAt(r, c int) rune {
	return e.row[r].runes[c]
}

func (e *editor) searchForward(startr, startc int, stext string) (fr, rc int) {
	if len(stext) == 0 {
		return -1, -1
	}
	s := []rune(stext)
	ss := 0

	c := startc
	r := startr
	for r < e.numrows {
		for c < e.row[r].size {
			rch := e.runeAt(r, c)
			if ss < len(s) && s[ss] == rch {
				ss++
			} else {
				if ss == len(s) {
					log.Println("found", r, c)
					return r, c
				}
				if s[ss] != rch {
					ss = 0
				}
			}
			c++
		}
		c = 0
		r++
		ss = 0
	}
	return -1, -1
}

func reverse(s string) []rune {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return r
}

func (e *editor) searchBackwards(startr, startc int, stext string) (fr, fc int) {
	if len(stext) == 0 {
		return -1, -1
	}
	s := reverse(stext)
	ss := 0

	c := startc
	r := startr
	//log.Printf("before lop r %d c %d\n", r, c)
	for r > -1 {
		log.Printf("r is %d\n", r)
		for c >= 0 && c < e.row[r].size {
			rch := e.runeAt(r, c)
			if ss < len(s) && s[ss] == rch {
				fr = r
				fc = c
				ss++
			} else if s[ss] != rch {
				ss = 0
			}

			if ss == len(s) {
				log.Println("found", r, c)
				return fr, fc
			}
			c--
		}
		if r > 0 {
			r--
			c = e.row[r].size - 1
		} else {
			break
		}
		ss = 0
	}
	return -1, -1
}

func (e *editor) editorFind() {
	query := ""
	startrow := e.rowoff + e.cy
	startcol := e.coloff + e.cx
	lastLineMatch := startrow /* Last line where a match was found. -1 for none. */
	lastColMatch := startcol  // last column where a match was found
	findDirection := 1        /* if 1 search next, if -1 search prev. */

	/* Save the cursor position in order to restore it later. */
	savedCx, savedCy := e.cx, e.cy
	savedColoff, savedRowoff := e.coloff, e.rowoff

	for {
		e.editorSetStatusMessage("Search: %s (Use Esc/Arrows/Enter)", query)
		e.editorRefreshScreen()
		ev := <-e.events
		if ev.Ch != 0 {
			ch := ev.Ch
			query = query + string(ch)
			lastLineMatch, lastColMatch = startrow, startcol
			findDirection = 1
		}
		if ev.Ch == 0 {
			switch ev.Key {
			case termbox.KeyTab:
				query = query + string('\t')
				lastLineMatch, lastColMatch = startrow, startcol
				findDirection = 1
			case termbox.KeySpace:
				query = query + string(' ')
				lastLineMatch, lastColMatch = startrow, startcol
				findDirection = 1
			case termbox.KeyEnter, termbox.KeyCtrlR:
				e.editorSetStatusMessage("")
				return
			case termbox.KeyCtrlC:
				e.editorSetStatusMessage("killed.")
				return
			case termbox.KeyBackspace2, termbox.KeyBackspace:
				if len(query) > 0 {
					query = query[:len(query)-1]
				} else {
					query = ""
				}
				lastLineMatch, lastColMatch = startrow, startcol
				findDirection = -1
			case termbox.KeyCtrlG, termbox.KeyEsc:
				e.cx = savedCx
				e.cy = savedCy
				e.coloff = savedColoff
				e.rowoff = savedRowoff
				e.editorSetStatusMessage("")
				return
			case termbox.KeyArrowDown, termbox.KeyArrowRight:
				findDirection = 1
			case termbox.KeyArrowLeft, termbox.KeyArrowUp:
				findDirection = -1
			default:
				e.editorSetStatusMessage("Search: %s (Use Esc/Arrows/Enter)", query)
				e.editorRefreshScreen()
				findDirection = 0
			}
		}
		if findDirection != 0 {
			currentLine, matchOffset := -1, -1
			if findDirection == 1 {
				currentLine, matchOffset = e.searchForward(lastLineMatch, lastColMatch, query)
			}
			if findDirection == -1 {
				currentLine, matchOffset = e.searchBackwards(lastLineMatch, lastColMatch, query)
			}
			//findDirection = 0

			if currentLine != -1 {
				lastLineMatch = currentLine
				lastColMatch = matchOffset
				e.cy = 0
				e.cx = matchOffset
				e.rowoff = currentLine
				e.coloff = 0
				/* Scroll horizontally as needed. */
				if e.cx > e.screencols {
					diff := e.cx - e.screencols
					e.cx -= diff
					e.coloff += diff
				}
			}
		}
	}
}

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

// NewEditor generates a new editor for use
func (e *editor) init() {
	e.done = false
	e.cx = 0
	e.cy = 0
	e.rowoff = 0
	e.coloff = 0
	e.numrows = 0
	e.row = nil
	e.dirty = false
	e.filename = ""
	e.screencols, e.screenrows = termbox.Size()
	e.screenrows -= 2 /* Get room for status bar. */
	e.quitTimes = 3
}

// Start runs an editor
func (z *Ziti) Start(filename string) error {
	z.ziti.init()
	e := z.ziti
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	e.editorOpen(filename)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	e.editorSetStatusMessage("HELP: Ctrl-S = save | Ctrl-Q = quit | Ctrl-F = find")

	e.events = make(chan termbox.Event, 20)
	go func() {
		for {
			e.events <- termbox.PollEvent()
		}
	}()
	for e.done != true {
		e.editorRefreshScreen()
		select {
		case ev := <-e.events:
			e.editorProcessEvent(ev)
			termbox.Flush()
		}
	}

}
