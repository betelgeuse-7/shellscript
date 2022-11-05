### ShellScript

A stack-based scripting language for Linux.

It includes common shell functionality like reading files, printing to standard out, etc.

This implementation uses syscalls to implement those features, instead of relying on Go standard library functions; just for the sake of learning how to use the ```syscall``` package of Go, and to gain a general understanding of system calls to the kernel.

_It has a very original name, I know._

#### Commands


#### Code snippets
-

#### Features

-
```
"hello" print       # print "hello" to stdout
"newfile.txt" newfile
"writetofile.txt" "write this" write

"newfile.txt" "otherfile.txt" read write

"*.txt" del     # delete all files ending with .txt

"file.go" "newname.go" rn #rename file.go to newname.go

"newname.go" read print         # print contents of newname.go to standard out.

```