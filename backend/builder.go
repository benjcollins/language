package backend

import "fmt"

func NewProgram() *Program {
	return &Program{[]*Value{}, []*Block{}}
}

func (program *Program) NewBlock() *Block {
	block := &Block{
		instructions:   []Instruction{},
		branch:         nil,
		liveOut:        map[*Value]struct{}{},
		previousBlocks: []*Block{},
		program:        program,
		name:           fmt.Sprintf("b%d", len(program.blocks)),
	}
	program.blocks = append(program.blocks, block)
	return block
}

func (program *Program) NewValue() *Value {
	name := fmt.Sprintf("v%d", len(program.values))
	val := &Value{
		defs:      []Instruction{},
		name:      name,
		interfere: map[*Value]struct{}{},
		register:  "",
		alive:     false,
		value:     0,
	}
	program.values = append(program.values, val)
	return val
}

func (block *Block) Binary(a, b *Value, op BinaryOp) *Value {
	dest := block.program.NewValue()

	if len(a.defs) == 1 && len(b.defs) == 1 {
		a, aok := a.defs[0].(*Constant)
		b, bok := b.defs[0].(*Constant)
		if aok && bok {
			inst := &Constant{dest, a.val + b.val}
			dest.defs = append(dest.defs, inst)
			block.instructions = append(block.instructions, inst)
		}
		return dest
	}

	inst := &Binary{a, b, dest, op}
	block.instructions = append(block.instructions, inst)
	dest.defs = append(dest.defs, inst)
	a.uses = append(a.uses, inst)
	b.uses = append(b.uses, inst)
	return dest
}

func (block *Block) Constant(value int) *Value {
	dest := block.program.NewValue()
	inst := &Constant{dest, value}
	block.instructions = append(block.instructions, inst)
	dest.defs = append(dest.defs, inst)
	return dest
}

func (block *Block) Copy(src, dest *Value) {
	inst := &Copy{src, dest}
	block.instructions = append(block.instructions, inst)
	src.uses = append(src.uses, inst)
	dest.defs = append(dest.defs, inst)
}

func (block *Block) Jump(target *Block) {
	target.previousBlocks = append(target.previousBlocks, block)
	block.branch = &Jump{target}
}

func (block *Block) ConditionalJump(a, b *Value, ifTrue, ifFalse *Block, cond JumpCondition) {
	ifTrue.previousBlocks = append(ifTrue.previousBlocks, block)
	ifFalse.previousBlocks = append(ifFalse.previousBlocks, block)
	branch := &ConditionalJump{a, b, ifTrue, ifFalse, cond}
	a.uses = append(a.uses, branch)
	b.uses = append(b.uses, branch)
	block.branch = branch
}

func (block *Block) JumpIfEqual(a, b *Value, ifTrue, ifFalse *Block) {
	block.ConditionalJump(a, b, ifTrue, ifFalse, Equal)
}

func (block *Block) JumpIfGreater(a, b *Value, ifTrue, ifFalse *Block) {
	block.ConditionalJump(a, b, ifTrue, ifFalse, Greater)
}

func (block *Block) JumpIfLessOrEqual(a, b *Value, ifTrue, ifFalse *Block) {
	block.ConditionalJump(a, b, ifTrue, ifFalse, LessOrEqual)
}

func (block *Block) Add(a, b *Value) *Value {
	return block.Binary(a, b, Add)
}

func (block *Block) Exit(val *Value) {
	branch := &Exit{val}
	val.uses = append(val.uses, branch)
	block.branch = branch
}

func (block *Block) Call(target, ret *Block) {
	target.previousBlocks = append(target.previousBlocks, block)
	ret.previousBlocks = append(ret.previousBlocks, target)
	block.branch = &Call{target, ret}
}

func (block *Block) Return() {
	block.branch = &Return{}
}
