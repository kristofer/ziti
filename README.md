# Ziti


Ziti is a small terminal/screen text editor in less than 2K lines of Go code. It uses Go's rune (unicode) machinery. Has multiple buffers. But still single window. 

### Usage: ziti `<filename>`

To build: clone repo into gopath;
 ```
  $ cd cmd
  $ go get -d ./...
  $ go build -o ziti
 ```

 Copy to someplace in your PATH

be sure you have https://godoc.org/github.com/nsf/termbox-go
so, you may need to `go get github.com/nsf/termbox-go`

### Key Commands:

#### Movement
* CTRL-Y: HELP
* Use ArrowKeys to move Cursor around.
* Home, End, and PageUp & PageDown should work
* CTRL-A: Move to beginning of current line
* CTRL-E: Move to end of current line
* on mac keyboards:
  * FN+ArrowUp: PageUp (screen full)
  * FN+ArrowDown: PageDown (screen full)

#### File/Buffer 
* CTRL-S: Save the file
* CTRL-Q: Quit the editor
* CTRL-F: Find string in file 
	(ESC to exit search mode, arrows to navigate to next/prev find)
* CTRL-N: Next Buffer
* CTRL-B: List all the buffers
* CTRL-O: (control oh) Open File into new buffer
* CTRL-W: kill buffer (not yet)

#### Cut/Copy/Paste & Deletion
* CTRL-Space: Set Mark
* CTRL-X: Cut region from Mark to Cursor into paste buffer
* CTRL-C: Copy region from Mark to Cursor into paste buffer
* CTRL-V: Paste copied/cut region into file at Cursor
_Once you've set the Mark, as you move the cursor,
you should be getting underlined text showing the current
selection/region._
* Delete: to delete a rune backward
* CTRL-K: killtoEndOfLine (once) removeLine (twice)


Setting the cursor with a mouse click should work. (and so,
it should work to set the selection. but hey, you MUST SetMark
for a selection to start... sorry, it's not a real mouse based editor.)
    
### Implementation Notes
Ziti was based on Kilo, a project by Salvatore Sanfilippo <antirez at gmail dot com> at  https://github.com/antirez/kilo.

It's a very simple editor, with kinda-"Mac-Emacs"-like key bindings. It uses `go get github.com/nsf/termbox-go" for simple termio junk.

The central data structure is an array of lines (type erow struct). Each line in the file has a struct, which contains an array of rune. (If you're not familiar with Go's _runes_, they are Go's unicode code points (or characters))

Multiple buffers, but no window splits. Two _mini modes_,  one for the search modal operations, and
one for opening files.

Notice the goroutine attached to events coming from termbox-go, that is pretty cool. Yet another real reason that Go routines are handy.

_Ziti was written in Go by K Younger and is released
under the BSD 2 clause license._
