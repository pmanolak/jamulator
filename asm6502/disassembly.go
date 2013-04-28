package asm6502

import (
	"io"
	"os"
	"fmt"
	"bufio"
	"strings"
)

type SourceWriter struct {
	program *Program
	writer *bufio.Writer
	Errors []string
}

type Renderer interface {
	Render(SourceWriter) error
}

func Disassemble(reader io.Reader) (*Program, error) {
	r := bufio.NewReader(reader)

	p := new(Program)
	p.Ast = new(ProgramAST)
	p.Ast.statements = make(StatementList, 0)

	for {
		opCode, err := r.ReadByte()
		if err == io.EOF { break }
		if err != nil { return nil, err }
		opCodeInfo := opCodeDataMap[opCode]
		switch opCodeInfo.addrMode {
		case nilAddr:
			i := new(DataStatement)
			i.dataList = make(DataList, 1)
			item := IntegerDataItem(opCode)
			i.dataList[0] = &item
			p.Ast.statements = append(p.Ast.statements, i)
		case absAddr:
		case absXAddr:
		case absYAddr:
		case immedAddr:
			i := new(ImmediateInstruction)
			i.OpName = opCodeInfo.opName
			v, err := r.ReadByte()
			if err != nil { return nil, err }
			i.Value = int(v)
			p.Ast.statements = append(p.Ast.statements, i)
		case impliedAddr:
			i := new(ImpliedInstruction)
			i.OpName = opCodeInfo.opName
			p.Ast.statements = append(p.Ast.statements, i)
		case indirectAddr:
		case xIndexIndirectAddr:
		case indirectYIndexAddr:
		case relativeAddr:
		case zeroPageAddr:
		case zeroXIndexAddr:
		case zeroYIndexAddr:

		}
	}

	return p, nil
}

func DisassembleFile(filename string) (*Program, error) {
	fd, err := os.Open(filename)
	if err != nil { return nil, err}

	p, err := Disassemble(fd)
	err2 := fd.Close()
	if err != nil { return nil, err }
	if err2 != nil { return nil, err2 }

	return p, nil
}

func (sw SourceWriter) Visit(n Node) {
	switch t := n.(type) {
	case Renderer:
		err := t.Render(sw)
		if err != nil {
			sw.Errors = append(sw.Errors, err.Error())
		}
	}
}

func (SourceWriter) VisitEnd(Node) {}

func (sw SourceWriter) Error() string {
	return strings.Join(sw.Errors, "\n")
}

func (i *ImmediateInstruction) Render(sw SourceWriter) error {
	_, err := sw.writer.WriteString(fmt.Sprintf("%s #$%02x\n", i.OpName, i.Value))
	return err
}

func (i *ImpliedInstruction) Render(sw SourceWriter) error {
	_, err := sw.writer.WriteString(fmt.Sprintf("%s\n", i.OpName))
	return err
}

func (s *DataStatement) Render(sw SourceWriter) error {
	_, err := sw.writer.WriteString("dc.b ")
	if err != nil { return err }
	for i, node := range(s.dataList) {
		switch t := node.(type) {
		case *StringDataItem:
			_, err = sw.writer.WriteString("\"")
			if err != nil { return err }
			_, err = sw.writer.WriteString(string(*t))
			if err != nil { return err }
			_, err = sw.writer.WriteString("\"")
			if err != nil { return err }
		case *IntegerDataItem:
			_, err = sw.writer.WriteString(fmt.Sprintf("#$%02x", int(*t)))
			if err != nil { return err }
		}
		if i < len(s.dataList) - 1 {
			sw.writer.WriteString(", ")
		}
	}
	_, err = sw.writer.WriteString("\n")
	return err
}

func (p *Program) WriteSource(writer io.Writer) error {
	w := bufio.NewWriter(writer)

	sw := SourceWriter{p, w, make([]string, 0)}
	p.Ast.Ast(sw)
	w.Flush()

	if len(sw.Errors) > 0 {
		return sw
	}
	return nil
}

func (p *Program) WriteSourceFile(filename string) error {
	fd, err := os.Create(filename)
	if err != nil { return err }
	err = p.WriteSource(fd)
	err2 := fd.Close()
	if err != nil { return err }
	if err2 != nil { return err2 }
	return nil
}
