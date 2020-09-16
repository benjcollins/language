package syntax

import "unicode/utf8"

type Position struct {
	line     int
	column   int
	leftover string
}

func startOfString(str string) Position {
	return Position{1, 1, str}
}

func (pos Position) next() Position {
	rune, size := utf8.DecodeRuneInString(pos.leftover)
	if rune == '\n' {
		return Position{pos.line + 1, 1, pos.leftover[size:]}
	}
	return Position{pos.line, pos.column + 1, pos.leftover[size:]}
}

func (pos Position) peek() rune {
	rune, _ := utf8.DecodeRuneInString(pos.leftover)
	return rune
}

func (pos Position) len() int {
	return len(pos.leftover)
}

func between(start, end Position) string {
	return start.leftover[:start.len()-end.len()]
}
