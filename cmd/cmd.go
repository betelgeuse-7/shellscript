package cmd

import (
	"github.com/betelgeuse-7/shellscript/pkg"
)

func Eval(input string) {
	l := pkg.NewLexer(input)
	p := pkg.NewParser(l)
	els := p.Parse()
	pkg.Eval(els)
}
