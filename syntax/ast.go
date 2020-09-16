package syntax

import "fmt"

type Expr interface{}

type Span struct {
	start, end Position
	expr       Expr
}

func (span Span) GetExpr() Expr {
	return span.expr
}

type Identifier struct {
	Ident string
}

type IntegerLiteral struct {
	Value string
}

type BooleanLiteral struct {
	Value string
}

type Binary struct {
	Op    BinaryOp
	Left  Span
	Right Span
}

type BinaryOp string

const (
	Add          BinaryOp = "+"
	SingleEquals BinaryOp = "="
	LessThan     BinaryOp = "<"
	Else         BinaryOp = "else"
	If           BinaryOp = "if"
	Func         BinaryOp = "fn"
	Call         BinaryOp = "call"
	While        BinaryOp = "while"
	Dot                   = "."
)

type Unary struct {
	Op   UnaryOp
	Expr Span
}

type UnaryOp string

const (
	Maybe UnaryOp = "?"
)

type Block struct {
	Statements []Span
}

type Tuple struct {
	Items []Span
}

type Struct struct {
	Block Span
}

func newBinary(left, right Span, op BinaryOp) Binary {
	return Binary{op, left, right}
}

func newInfix(left, right Span, op BinaryOp) Span {
	return Span{left.start, right.end, Binary{op, left, right}}
}

func newBlock(left, right Span) Span {
	block := Block{[]Span{left, right}}
	if left, ok := left.expr.(Block); ok {
		block.Statements = append(left.Statements, right)
	}
	return Span{left.start, right.end, block}
}

func newTuple(left, right Span) Span {
	tuple := Tuple{[]Span{left, right}}
	if left, ok := left.expr.(Tuple); ok {
		tuple.Items = append(left.Items, right)
	}
	return Span{left.start, right.end, tuple}
}

func (span Span) String() string {
	return formatAST(span, "", 0)
}

func formatAST(span Span, indent string, line int) string {
	s := ""
	if span.start.line == line {
		s += fmt.Sprintf("    | %s", indent)
	} else {
		s += fmt.Sprintf("%3d | %s", span.start.line, indent)
	}
	line = span.start.line
	indent += "   "
	switch expr := span.expr.(type) {
	case Identifier:
		s += fmt.Sprintf("ident: '%s'\n", expr.Ident)
	case IntegerLiteral:
		s += fmt.Sprintf("integer: '%s'\n", expr.Value)
	case BooleanLiteral:
		s += fmt.Sprintf("boolean: '%s'\n", expr.Value)
	case Binary:
		s += fmt.Sprintf("binary '%s':\n", expr.Op)
		s += formatAST(expr.Left, indent, line)
		line = expr.Left.start.line
		s += formatAST(expr.Right, indent, line)
	case Unary:
		s += fmt.Sprintf("unary '%s':\n", expr.Op)
		s += formatAST(expr.Expr, indent, line)
	case Block:
		s += fmt.Sprintf("block:\n")
		for _, statement := range expr.Statements {
			s += formatAST(statement, indent, line)
			line = statement.start.line
		}
	case Tuple:
		s += fmt.Sprintf("tuple:\n")
		for _, item := range expr.Items {
			s += formatAST(item, indent, line)
			line = item.start.line
		}
	case Struct:
		s += fmt.Sprintf("struct:\n")
		s += formatAST(expr.Block, indent, line)
	default:
		s += "error!"
	}
	return s
}
