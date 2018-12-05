# Ziti


Ziti is a small terminal/screen text editor in less than 1K lines of Go code. It uses Go's rune (unicode) machinery.

### Usage: ziti `<filename>`

To build: clone repo into gopath;
 ```
  $ go get -d ./...
  $ go build -o ziti
 ```

be sure you have https://godoc.org/github.com/nsf/termbox-go
so, you may need to `go get github.com/nsf/termbox-go`

###Keys:

    ArrowKeys to move around.

    CTRL-S: Save
    CTRL-Q: Quit
    CTRL-F: Find string in file (ESC to exit search mode, arrows to navigate)

    CTRL-Space: Set Mark
    CTRL-X: Cut region from Mark to Cursor into paste buffer
    CTRL-C: Copy region from Mark to Cursor into paste buffer
    CTRL-V: Paste copied/cut region into file at Cursor

    Use Arrows to move, Home, End, and PageUp & PageDown
    CTRL-A: Move to beginning of current line
    CTRL-E: Move to end of current line

    Delete to delete a rune backward

    on mac keyboards:
    FN+ArrowUp: PageUp (screen full)
    FN+ArrowDown: PageDown (screen full)
    

Ziti was based on Kilo, a project by Salvatore Sanfilippo <antirez at gmail dot com> at  https://github.com/antirez/kilo.

It's a very simple editor, with kinda-"Mac-Emacs"-like key bindings. It uses `go get github.com/nsf/termbox-go" for simple termio junk.

The central data structure is an array of lines (type erow struct). Each line in the file has a struct, which contains an array of rune. (If you're not familiar with Go's _runes_, they are Go's unicode code points (or characters))

Single file, no buffers, no window splits. One _mini mode_, for the search modal operations.

But the goroutine attached to events coming from termbox-go is pretty cool. Yet another real reason that Go routines
are handy.

Ziti was written in Go by K Younger and is released
under the BSD 2 clause license.
