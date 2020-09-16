package backend

import "fmt"

func X86(program *Program, entry *Block) string {
	finishedBlocks := map[*Block]struct{}{}
	str := ""
	queue := []*Block{entry}

	for len(queue) > 0 {
		block := queue[len(queue)-1]
		finishedBlocks[block] = struct{}{}
		queue = queue[:len(queue)-1]
		str += fmt.Sprintf("%s:\n", block.name)

		for _, inst := range block.instructions {
			switch inst := inst.(type) {
			case *Constant:
				str += fmt.Sprintf("  mov $%d, %s\n", inst.val, inst.dest.register)
			case *Binary:
				if inst.a.register == inst.dest.register {
					str += fmt.Sprintf("  add %s, %s\n", inst.b.register, inst.a.register)
				} else if inst.b.register == inst.dest.register {
					str += fmt.Sprintf("  add %s, %s\n", inst.a.register, inst.b.register)
				} else {
					str += fmt.Sprintf("  mov %s, %s\n", inst.a.register, inst.dest.register)
					str += fmt.Sprintf("  add %s, %s\n", inst.b.register, inst.dest.register)
				}
			case *Copy:
				if inst.src.register != inst.dest.register {
					str += fmt.Sprintf("  mov %s, %s\n", inst.src.register, inst.dest.register)
				}
			}
		}
		switch branch := block.branch.(type) {
		case *Jump:
			if _, finished := finishedBlocks[branch.target]; finished {
				str += fmt.Sprintf("  jmp %s\n", branch.target.name)
			} else {
				queue = append(queue, branch.target)
			}
		case *ConditionalJump:
			str += fmt.Sprintf("  cmp %s, %s\n", branch.b.register, branch.a.register)
			switch branch.cond {
			case Equal:
				str += fmt.Sprintf("  je %s\n", branch.ifTrue.name)
			case Greater:
				str += fmt.Sprintf("  ja %s\n", branch.ifTrue.name)
			case LessOrEqual:
				str += fmt.Sprintf("  jle %s\n", branch.ifTrue.name)
			}
			if _, finished := finishedBlocks[branch.ifTrue]; !finished {
				queue = append(queue, branch.ifTrue)
			}
			if _, finished := finishedBlocks[branch.ifFalse]; finished {
				str += fmt.Sprintf("  jmp %s\n", branch.ifFalse.name)
			} else {
				queue = append(queue, branch.ifFalse)
			}
		case *Exit:
			str += fmt.Sprintf("  mov %s, %%edi\n", branch.val.register)
			str += "  mov $60, %eax\n"
			str += "  syscall\n"
		}
	}

	return str
}
