package main

import "github.com/betelgeuse-7/shellscript/cmd"

func main() {
	//input := `"hello.txt" read print`
	//input := `     `
	//input := `"hello2.txt" "write this to hello2.txt" write`
	input := `"Hello" print "Hello 2" print "hello.txt" read print "Finish" print`
	cmd.Eval(input)
}
