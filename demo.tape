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

# 4. Navigate to a root file and preview it
Type "jjj"
Sleep 500ms
Type "l"
Sleep 2s

# 5. Search in preview pane
Type "/"
Sleep 500ms
Type "make"
Sleep 1.5s
Escape
Sleep 500ms

# 6. Scroll preview pane
Type "jjjjj"
Sleep 1.5s
Type "kkk"
Sleep 500ms
Type "h"
Sleep 500ms

# 7. Search in explorer pane
Type "/"
Sleep 500ms
Type "go"
Sleep 1.5s
Escape
Sleep 500ms

# 8. Toggle hidden files
Type "i"
Sleep 1.5s
Type "i"
Sleep 1s

# 9. Collapse / expand the file list pane
Type "	"
Sleep 2s
Type "	"
Sleep 1s

# Exit
Type "q"
