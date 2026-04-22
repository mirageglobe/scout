Output demo.gif

Set FontSize 20
Set Width 1200
Set Height 600
Set Padding 20
Set WindowBar Colorful
Set Theme "Dracula"

# Start Scout
Type "./scout"
Enter
Sleep 2s

# 1. Navigate — show git badges and file list
Type "jj"
Sleep 500ms
Type "jj"
Sleep 500ms
Type "g"
Sleep 500ms

# 2. Help overlay
Type "?"
Sleep 3s
Type "?"
Sleep 1s

# 3. Cycle themes
Type "t"
Sleep 1s
Type "t"
Sleep 1s
Type "t"
Sleep 1s
Type "t"
Sleep 1.5s

# 4. Enter cmd/ directory and preview main.go with syntax highlighting
Type "j"
Sleep 500ms
Type "l"
Sleep 1s
Type "j"
Sleep 2s

# 5. Search in explorer pane
Type "h"
Sleep 500ms
Type "/"
Sleep 500ms
Type "main"
Sleep 1.5s
Escape
Sleep 500ms

# 6. Focus preview pane and search within file
Type "l"
Sleep 500ms
Type "/"
Sleep 500ms
Type "func"
Sleep 1.5s
Escape
Sleep 500ms

# 7. Scroll preview pane
Type "jjjjj"
Sleep 1.5s
Type "kkk"
Sleep 500ms
Type "h"
Sleep 500ms

# 8. Navigate back to root
Type "h"
Sleep 1s

# 9. Toggle hidden files
Type "i"
Sleep 1.5s
Type "i"
Sleep 1s

# 10. Collapse / expand the file list pane
Type "	"
Sleep 2s
Type "	"
Sleep 1s

# Exit
Type "q"
