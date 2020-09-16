package backend

import "fmt"

func Execute(block *Block) {
	stack := []*Block{}
	for {
		for _, inst := range block.instructions {
			switch inst := inst.(type) {
			case *Constant:
				inst.dest.value = inst.val

			case *Binary:
				switch inst.op {
				case Add:
					inst.dest.value = inst.a.value + inst.b.value
				}

			case *Copy:
				inst.dest.value = inst.src.value
			}
		}
		switch branch := block.branch.(type) {
		case *Jump:
			block = branch.target

		case *ConditionalJump:
			condition := false
			switch branch.cond {
			case LessOrEqual:
				condition = branch.a.value <= branch.b.value
			case Greater:
				condition = branch.a.value > branch.b.value
			case Equal:
				condition = branch.a.value == branch.b.value
			}
			if condition {
				block = branch.ifTrue
			} else {
				block = branch.ifFalse
			}

		case *Exit:
			fmt.Println(branch.val.value)
			return

		case *Call:
			stack = append(stack, branch.ret)
			block = branch.target

		case *Return:
			block = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
		}
	}
}
