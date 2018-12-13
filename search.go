package ziti

import (
	"log"

	termbox "github.com/nsf/termbox-go"
)

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
	startrow := e.point.ro + e.point.r
	startcol := e.point.co + e.point.c
	lastLineMatch := startrow /* Last line where a match was found. -1 for none. */
	lastColMatch := startcol  // last column where a match was found
	findDirection := 1        /* if 1 search next, if -1 search prev. */

	/* Save the cursor position in order to restore it later. */
	savedCx, savedCy := e.point.c, e.point.r
	savedColoff, savedRowoff := e.point.co, e.point.ro

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
				e.point.c, e.point.r = savedCx, savedCy
				e.point.co, e.point.ro = savedColoff, savedRowoff
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
				e.point.r = 0
				e.point.c = matchOffset
				e.point.ro = currentLine
				e.point.co = 0
				/* Scroll horizontally as needed. */
				if e.point.c > e.screencols {
					diff := e.point.c - e.screencols
					e.point.c -= diff
					e.point.co += diff
				}
			}
		}
	}
}
