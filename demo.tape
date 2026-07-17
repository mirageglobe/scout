# privacy: record from a neutral $HOME so scout's header shows ~/scout,
# not your real checkout path (this gif is public).
# see CONTRIBUTING.md "regenerating the demo".
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

# 1. Navigate — show git badges, file sizes, changed-file count
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
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
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Right
Sleep 2s

# 5. Search in preview pane
Type "/"
Sleep 500ms
Type "make"
Sleep 1.5s
Enter
Sleep 500ms
Type "n"
Sleep 500ms
Type "n"
Sleep 500ms
Escape
Sleep 500ms

# 6. Scroll preview pane
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 1.5s
Up
Sleep 300ms
Up
Sleep 300ms
Up
Sleep 500ms
Left
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

# 9. Root lock (new in v0.6.0)
Type "l"
Sleep 1.5s
Type "l"
Sleep 1s

# 10. Collapse / expand the file list pane
Tab
Sleep 2s
Tab
Sleep 1s

# Exit
Type "q"
