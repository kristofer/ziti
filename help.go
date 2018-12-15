package ziti

import (
	"bufio"
	"log"
	"strings"
)

const helptext = ` ** ZITI: Help on keyboard commands

 CTRL-Y: HELP
 Use ArrowKeys to move Cursor around.

 CTRL-S: Save the file
 CTRL-Q: Quit the editor
 CTRL-F: Find string in file 
	(ESC to exit search mode, arrows to navigate to next/prev find)
 CTRL-N: Next Buffer
 CTRL-B: List all the buffers
 CTRL-O: (control oh) Open File into new buffer
 CTRL-W: kill buffer

 CTRL-Space: Set Mark
 CTRL-X: Cut region from Mark to Cursor into paste buffer
 CTRL-C: Copy region from Mark to Cursor into paste buffer
 CTRL-V: Paste copied/cut region into file at Cursor
Once you've set the Mark, as you move the cursor,
you should be getting underlined text showing the current
selection/region.

 Use Arrows to move, Home, End, and PageUp & PageDown should work
 CTRL-A: Move to beginning of current line
 CTRL-E: Move to end of current line

 Delete: to delete a rune backward
 CTRL-K: killtoEndOfLine (once) removeLine (twice)

 on mac keyboards:
 FN+ArrowUp: PageUp (screen full)
 FN+ArrowDown: PageDown (screen full)

Setting the cursor with a mouse click should work. (and so,
it should work to set the selection. but hey, you MUST SetMark
for a selection to start... sorry, it's not a real mouse based editor.)
`

func (e *editor) loadHelp() error {
	nb := &buffer{}
	e.buffers = append(e.buffers, nb)
	e.cb = nb
	e.cb.filename = "*Ziti Help*"
	scanner := bufio.NewScanner(strings.NewReader(helptext))
	for scanner.Scan() {
		// does the line contain the newline?
		line := scanner.Text()
		e.editorInsertRow(e.cb.numrows, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	e.cb.dirty = false
	e.cb.readonly = true
	return nil
}
