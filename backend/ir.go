package backend

import "fmt"

type Value struct {
	defs      []Instruction
	uses      []Instruction
	name      string
	interfere map[*Value]struct{}
	register  string
	alive     bool
	value     int
}

func (value Value) String() string {
	return value.name
}

type BinaryOp string

const (
	Add BinaryOp = "add"
)

type Binary struct {
	a, b, dest *Value
	op         BinaryOp
}

type Constant struct {
	dest *Value
	val  int
}

type Copy struct {
	src, dest *Value
}

type Instruction interface{}

type Branch interface{}

type Program struct {
	values []*Value
	blocks []*Block
}

type Block struct {
	instructions   []Instruction
	branch         Branch
	liveOut        map[*Value]struct{}
	previousBlocks []*Block
	program        *Program
	name           string
}

type Exit struct {
	val *Value
}

type Jump struct {
	target *Block
}

type JumpCondition string

const (
	Equal       JumpCondition = "equal"
	LessOrEqual JumpCondition = "lessOrEqual"
	Greater     JumpCondition = "greater"
)

type ConditionalJump struct {
	a, b            *Value
	ifTrue, ifFalse *Block
	cond            JumpCondition
}

type Call struct {
	target, ret *Block
}

type Return struct{}

func IrToStr(program *Program) string {
	str := ""
	for _, block := range program.blocks {
		str += fmt.Sprintf("%s {\n", block.name)
		for _, inst := range block.instructions {
			switch inst := inst.(type) {
			case *Constant:
				str += fmt.Sprintf("  %s = %d\n", inst.dest.name, inst.val)
			case *Binary:
				str += fmt.Sprintf("  %s = %s + %s\n", inst.dest.name, inst.a.name, inst.b.name)
			case *Copy:
				str += fmt.Sprintf("  %s = %s\n", inst.dest.name, inst.src.name)
			}
		}
		switch branch := block.branch.(type) {
		case *Jump:
			str += fmt.Sprintf("  goto %s\n", branch.target.name)
		case *ConditionalJump:
			str += fmt.Sprintf("  goto %s if %s < %s else goto %s\n", branch.ifTrue.name, branch.a.name, branch.b.name, branch.ifFalse.name)
		case *Exit:
			str += fmt.Sprintf("  exit(%s)\n", branch.val)
		case *Call:
			str += fmt.Sprintf("  %s()\n", branch.target.name)
			str += fmt.Sprintf("  goto %s\n", branch.ret.name)
		case *Return:
			str += fmt.Sprintf("  return\n")
		}
		str += "}\n\n"
	}
	return str
}

func (block *Block) SetName(name string) {
	block.name = name
}
