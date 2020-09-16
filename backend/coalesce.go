package backend

func CoalesceValues(before, after *Value) {
	for _, inst := range before.defs {
		switch inst := inst.(type) {
		case *Constant:
			inst.dest = after
		case *Binary:
			inst.dest = after
		case *Copy:
			inst.dest = after
		}
	}
	for _, inst := range before.uses {
		switch inst := inst.(type) {
		case *Binary:
			inst.a = ReplaceValue(inst.a, before, after)
			inst.b = ReplaceValue(inst.b, before, after)
		case *Copy:
			inst.src = ReplaceValue(inst.src, before, after)
		case *Exit:
			inst.val = ReplaceValue(inst.val, before, after)
		case *ConditionalJump:
			inst.a = ReplaceValue(inst.a, before, after)
			inst.b = ReplaceValue(inst.b, before, after)
		}
	}
	for val := range before.interfere {
		after.interfere[val] = struct{}{}
		val.interfere[after] = struct{}{}
	}
}

func ReplaceValue(current, before, after *Value) *Value {
	if current == before {
		return after
	}
	return current
}

func CoalesceCopies(program *Program) {
	for _, block := range program.blocks {
		insts := []Instruction{}
		for _, inst := range block.instructions {
			switch inst := inst.(type) {
			case *Copy:
				if _, interfere := inst.src.interfere[inst.dest]; !interfere {
					inst.dest.defs = append(inst.dest.defs, inst.src.defs...)
					inst.dest.uses = append(inst.dest.uses, inst.src.uses...)
					CoalesceValues(inst.src, inst.dest)
				} else {
					insts = append(insts, inst)
				}
			default:
				insts = append(insts, inst)
			}
		}
		block.instructions = insts
	}
}

func CoalesceBinary(program *Program) {
	for _, block := range program.blocks {
		insts := []Instruction{}
		for _, inst := range block.instructions {
			switch inst := inst.(type) {
			case *Binary:
				if _, interfere := inst.a.interfere[inst.dest]; !interfere {
					CoalesceValues(inst.a, inst.dest)
				} else if _, interfere := inst.b.interfere[inst.dest]; !interfere {
					CoalesceValues(inst.b, inst.dest)
				}
				insts = append(insts, inst)
			default:
				insts = append(insts, inst)
			}
		}
		block.instructions = insts
	}
}
