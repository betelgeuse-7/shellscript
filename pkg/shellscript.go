package pkg

import (
	"fmt"
	"log"
	"os"
	"syscall"
)

type TokenType uint32

const eof = 0

const (
	EOF TokenType = iota
	NEWLINE
	ILLEGAL
	STRING
	PRINT
	NEWFILE
	WRITE
	READ
)

func (tt TokenType) String() string {
	return map[TokenType]string{
		EOF:     "EndOfFile",
		NEWLINE: "Newline",
		ILLEGAL: "Illegal",
		STRING:  "String",
		PRINT:   "Print",
		NEWFILE: "Newfile",
		WRITE:   "Write",
		READ:    "Read",
	}[tt]
}

type Token struct {
	Typ     TokenType
	Literal string
	Line    int
}

func (t Token) String() string {
	return fmt.Sprintf("(%s, %s)", t.Typ, t.Literal)
}

func NewToken(typ TokenType, lit string, line int) Token {
	return Token{Typ: typ, Literal: lit, Line: line}
}

type Lexer struct {
	input                           string
	inputLen, counter, line         int
	char                            byte
	shouldIncrementLineAfterThisOne bool
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, counter: 0, line: 1}
	l.inputLen = len(l.input)
	if l.inputLen > 0 {
		l.char = l.input[l.counter]
	}
	return l
}

func (l *Lexer) next() {
	/* FIXME */
	l.counter++
	if l.counter-1 == l.inputLen-1 {
		l.char = eof
		return
	}
	l.char = l.input[l.counter]
	if l.shouldIncrementLineAfterThisOne {
		l.line++
		l.shouldIncrementLineAfterThisOne = false
	}
	if l.char == '\n' {
		l.shouldIncrementLineAfterThisOne = true
	}
}

func isLetter(char byte) bool {
	return char <= 'Z' && char >= 'A' || char >= 'a' && char <= 'z'
}

func isWs(char byte) bool {
	return char == '\t' || char == '\r' || char == ' '
}

func (l *Lexer) errorf(msgf string, args ...interface{}) {
	log.Fatalf(fmt.Sprintf("Lexer: %s\t\tAT LINE %d\n", msgf, l.line), args...)
}

func (l *Lexer) Lex() Token {
	if len(l.input) < 1 || l.char == eof {
		return NewToken(EOF, "EndOfFile", l.line)
	}
	for isWs(l.char) {
		l.next()
	}
	if isLetter(l.char) {
		start := l.counter
		for isLetter(l.char) {
			l.next()
		}
		lit := l.input[start:l.counter]
		switch lit {
		case "print":
			return NewToken(PRINT, lit, l.line)
		case "newfile":
			return NewToken(NEWFILE, lit, l.line)
		case "write":
			return NewToken(WRITE, lit, l.line)
		case "read":
			return NewToken(READ, lit, l.line)
		default:
			return NewToken(ILLEGAL, lit, l.line)
		}
	}
	switch l.char {
	case '\n':
		fmt.Println("newline")
		l.next()
		return NewToken(NEWLINE, "Newline", l.line)
	case '"':
		l.next()
		start := l.counter
		for l.char != '"' {
			if l.char == eof {
				l.errorf("unclosed string literal")
			}
			l.next()
		}
		lit := l.input[start:l.counter]
		l.next()
		return NewToken(STRING, lit, l.line)
	}
	l.next()
	return NewToken(ILLEGAL, string(l.char), l.line)
}

type Element interface {
	el()
}

type PrintCommand struct {
	Line int
	Arg  string
}

func (PrintCommand) el() {}

type NewfileCommand struct {
	Line     int
	Filename string
}

func (NewfileCommand) el() {}

type WriteCommand struct {
	Line              int
	Content, Filename string
}

func (WriteCommand) el() {}

type ReadCommand struct {
	Line     int
	Filename string
}

func (ReadCommand) el() {}

type StringLiteral struct {
	Literal string
	Line    int
}

func (StringLiteral) el() {}

type nodeStack struct {
	els []Element
}

func (n *nodeStack) pop() Element {
	if len(n.els) > 0 {
		popped := n.els[len(n.els)-1]
		n.els = n.els[:len(n.els)-1]
		return popped
	}
	return nil
}

func (n *nodeStack) push(el Element) {
	n.els = append(n.els, el)
}

type Parser struct {
	lexer *Lexer
	stack *nodeStack
}

func NewParser(l *Lexer) *Parser {
	e := &Parser{lexer: l}
	e.stack = &nodeStack{}
	return e
}

func (*Parser) errorf(line int, msgf string, args ...interface{}) {
	log.Fatalf(fmt.Sprintf("Evaluator: %s\t\tAT LINE %d\n", msgf, line), args...)
}

func (p *Parser) errif(cond bool, line int, errmsgf string, args ...interface{}) {
	if cond {
		p.errorf(line, errmsgf, args...)
	}
}

func elementIsNil(el Element) bool {
	return el == nil
}

func notStringLit(el Element) bool {
	_, ok := el.(StringLiteral)
	return !(ok)
}

func expandReadCommand(p *Parser, poppedEl ReadCommand) {
	// read file contents and push it as a StringLiteral to the stack
	filename := p.stack.pop()
	p.errif(elementIsNil(filename), poppedEl.Line, "missing file name argument for 'read' command")
	p.errif(notStringLit(filename), poppedEl.Line, "invalid file name argument for 'read'")
	fileContents, err := do_Read(filename.(StringLiteral).Literal)
	p.errif(err != nil, poppedEl.Line, fmt.Errorf("%w", err).Error())
	p.stack.push(StringLiteral{Literal: fileContents, Line: poppedEl.Line})
}

func (p *Parser) Parse() (els []Element) {
loop:
	for {
		tok := p.lexer.Lex()
		switch tok.Typ {
		case EOF:
			break loop
		case ILLEGAL:
			p.errorf(tok.Line, "illegal '%s'", tok.Literal)
		case NEWLINE:
			continue loop
		case STRING:
			p.stack.push(StringLiteral{Literal: tok.Literal, Line: tok.Line})
		case PRINT:
			p.stack.push(PrintCommand{Line: tok.Line})
		case NEWFILE:
			p.stack.push(NewfileCommand{Line: tok.Line})
		case WRITE:
			p.stack.push(WriteCommand{Line: tok.Line})
		case READ:
			p.stack.push(ReadCommand{Line: tok.Line})
		}
	}
	for {
		popped := p.stack.pop()
		if popped == nil {
			break
		}
		switch poppedEl := popped.(type) {
		case ReadCommand:
			expandReadCommand(p, poppedEl)
		case PrintCommand:
			arg1 := p.stack.pop()
			p.errif(elementIsNil(arg1), poppedEl.Line, "missing argument for 'print' command")
			if rc, ok := arg1.(ReadCommand); ok {
				expandReadCommand(p, rc)
				arg1 = p.stack.pop()
			}
			p.errif(notStringLit(arg1), poppedEl.Line, "invalid argument for 'print'")
			poppedEl.Arg = arg1.(StringLiteral).Literal
			els = append(els, poppedEl)
		case NewfileCommand:
			arg1 := p.stack.pop()
			p.errif(elementIsNil(arg1), poppedEl.Line, "missing argument for 'newfile' command")
			p.errif(notStringLit(arg1), poppedEl.Line, "invalid argument for 'newfile'")
			poppedEl.Filename = arg1.(StringLiteral).Literal
			els = append(els, poppedEl)
		case WriteCommand:
			content := p.stack.pop()
			filename := p.stack.pop()
			p.errif(elementIsNil(content), poppedEl.Line, "missing content argument for 'write' command")
			p.errif(elementIsNil(filename), poppedEl.Line, "missing file name argument for 'write' command")
			p.errif(notStringLit(content) || notStringLit(filename), poppedEl.Line, "invalid arguments for 'write'")
			poppedEl.Content = content.(StringLiteral).Literal
			poppedEl.Filename = filename.(StringLiteral).Literal
			els = append(els, poppedEl)
		case StringLiteral:
			p.errorf(poppedEl.Line, "lonely string literal '%s'", poppedEl.Literal)
		}
	}
	return
}

func Eval(els []Element) {
	for _, e := range els {
		switch c := e.(type) {
		case PrintCommand:
			if err := do_Print(c.Arg); err != nil {
				log.Fatalln(err)
			}
		case WriteCommand:
			if err := do_Write(c.Filename, c.Content); err != nil {
				log.Fatalln(err)
			}
		case NewfileCommand:
			if err := do_Write(c.Filename, ""); err != nil {
				log.Fatalln(err)
			}
		}
	}
}

func do_Print(s string) error {
	bx := []byte(s)
	newline := byte(10)
	bx = append(bx, newline)
	_, err := syscall.Write(int(os.Stdout.Fd()), bx)
	if err != nil {
		return fmt.Errorf("print: %w", err)
	}
	return nil
}

// fd, size, err
func fileDescriptorAndSizeOf(filename string) (int, int64, error) {
	f, err := os.Open(filename)
	if err != nil {
		return -1, -1, err
	}
	fsInfo, err := f.Stat()
	if err != nil {
		return -1, -1, err
	}
	size := fsInfo.Size()
	if err != nil {
		if os.IsNotExist(err) {
			return -1, -1, fmt.Errorf("file '%s' does not exist", filename)
		}
	}
	return int(f.Fd()), size, nil
}

func do_Read(filename string) (string, error) {
	fd, size, err := fileDescriptorAndSizeOf(filename)
	defer syscall.Close(fd)
	if err != nil {
		return "", err
	}
	var buf = make([]byte, size)
	_, err = syscall.Read(fd, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func do_Write(filename, content string) error {
	fd, _, err := fileDescriptorAndSizeOf(filename)
	defer syscall.Close(fd)
	if err != nil {
		return err
	}
	_, err = syscall.Write(fd, []byte(content))
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}
