package backend

func MarkUsedValue(val *Value) {
	if !val.alive {
		val.alive = true
		for _, inst := range val.defs {
			switch inst := inst.(type) {
			case *Binary:
				MarkUsedValue(inst.a)
				MarkUsedValue(inst.b)
			case *Copy:
				MarkUsedValue(inst.src)
			}
		}
	}
}

func MarkUsedValues(program *Program) {
	for _, block := range program.blocks {
		switch branch := block.branch.(type) {
		case *Exit:
			MarkUsedValue(branch.val)
		case *ConditionalJump:
			MarkUsedValue(branch.a)
			MarkUsedValue(branch.b)
		}
	}
}

func RemoveDeadCode(program *Program) {
	for _, block := range program.blocks {
		insts := []Instruction{}
		for _, inst := range block.instructions {
			switch inst := inst.(type) {
			case *Constant:
				if inst.dest.alive {
					insts = append(insts, inst)
				}
			case *Copy:
				if inst.dest.alive {
					insts = append(insts, inst)
				}
			case *Binary:
				if inst.dest.alive {
					insts = append(insts, inst)
				}
			}
		}
		block.instructions = insts
	}
}
