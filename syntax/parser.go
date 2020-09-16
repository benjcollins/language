package syntax

import (
	"fmt"
	"unicode"
)

type precedence int

const (
	block          precedence = iota
	statement      precedence = iota
	tuple          precedence = iota
	expr           precedence = iota
	boolean        precedence = iota
	comparison     precedence = iota
	subtraction    precedence = iota
	addition       precedence = iota
	multiplication precedence = iota
	divison        precedence = iota
	indicies       precedence = iota
	dot            precedence = iota
	literal        precedence = iota
	bracket        precedence = iota
)

func skipSpaces(start Position) (end Position) {
	end = start
	for unicode.IsSpace(end.peek()) {
		end = end.next()
	}
	return
}

func Parse(source string) (Span, []error) {
	parser := &parser{}
	expr := parser.parse(startOfString(source), block)
	return expr, parser.errors
}

type parser struct {
	errors []error
}

type parseError struct {
	pos Position
	msg string
}

func (err parseError) Error() string {
	return fmt.Sprintf("(%d, %d) %s", err.pos.line, err.pos.column, err.msg)
}

func (parser *parser) throw(pos Position, msg string) {
	parser.errors = append(parser.errors, parseError{pos, msg})
}

func (parser *parser) parseBracket(start Position, close rune, prec precedence) Span {
	left := parser.parse(start.next(), prec)
	pos := skipSpaces(left.end)
	if pos.peek() == close {
		return Span{start, pos.next(), left.expr}
	}
	parser.throw(pos, "missing close bracket")
	return Span{start, pos, nil}
}

func (parser *parser) parse(start Position, prec precedence) (left Span) {
	start = skipSpaces(start)
	switch {
	case unicode.IsDigit(start.peek()) && prec <= literal:
		end := start.next()
		for unicode.IsDigit(end.peek()) {
			end = end.next()
		}
		left = Span{start, end, IntegerLiteral{between(start, end)}}

	case start.peek() == '(' && prec <= bracket:
		return parser.parseBracket(start, ')', tuple)

	case start.peek() == '{' && prec <= bracket:
		return parser.parseBracket(start, '}', block)

	case unicode.IsLetter(start.peek()) && prec <= literal:
		end := start.next()
		for unicode.IsLetter(end.peek()) || unicode.IsDigit(end.peek()) {
			end = end.next()
		}
		switch between(start, end) {

		case "if":
			cond := parser.parse(end, bracket)
			conc := parser.parse(cond.end, expr)
			left = Span{start, conc.end, newBinary(cond, conc, If)}

		case "while":
			cond := parser.parse(end, bracket)
			body := parser.parse(cond.end, expr)
			left = Span{start, body.end, newBinary(cond, body, While)}

		case "struct":
			block := parser.parse(end, bracket)
			left = Span{start, block.end, Struct{block}}

		case "fn":
			param := parser.parse(end, bracket)
			body := parser.parse(param.end, expr)
			left = Span{start, body.end, newBinary(param, body, Func)}

		case "true", "false":
			left = Span{start, end, BooleanLiteral{between(start, end)}}

		default:
			left = Span{start, end, Identifier{between(start, end)}}
		}
	default:
		parser.throw(start, "no value")
		return Span{start, start.next(), nil}
	}

	for {
		start := skipSpaces(left.end)
		keywordEnd := start
		for unicode.IsLetter(keywordEnd.peek()) {
			keywordEnd = keywordEnd.next()
		}

		switch {
		case start.peek() == '+' && prec <= addition:
			left = newInfix(left, parser.parse(start.next(), addition), Add)

		case start.peek() == '?' && prec <= expr:
			left = Span{left.start, start.next(), Unary{Maybe, left}}

		case start.peek() == '.' && prec <= dot:
			left = newInfix(left, parser.parse(start.next(), dot), Dot)

		case start.peek() == '=' && prec <= statement:
			left = newInfix(left, parser.parse(start.next(), statement), SingleEquals)

		case start.peek() == '<' && prec <= comparison:
			left = newInfix(left, parser.parse(start.next(), comparison), LessThan)

		case between(start, keywordEnd) == "else" && prec < expr:
			left = newInfix(left, parser.parse(keywordEnd, expr), Else)

		case start.peek() == '(' && prec <= bracket:
			left = newInfix(left, parser.parse(start, bracket), Call)

		case start.peek() == ',' && prec <= tuple:
			left = newTuple(left, parser.parse(start.next(), tuple))

		case start.peek() != ')' && start.peek() != '}' && start.len() > 0 && prec <= block:
			left = newBlock(left, parser.parse(start, statement))

		default:
			return left
		}
	}
}
