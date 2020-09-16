package backend

func LivenessAnalysis(exit *Block, program *Program) {
	for _, block := range program.blocks {
		block.liveOut = map[*Value]struct{}{}
	}
	livenessAnalysis(exit)
}

func KillValue(liveIn map[*Value]struct{}, val *Value) {
	delete(liveIn, val)
}

func ReviveValue(liveIn map[*Value]struct{}, value *Value) {
	if _, prs := liveIn[value]; !prs {
		for val := range liveIn {
			InterfereValues(value, val)
		}
		liveIn[value] = struct{}{}
	}
}

func InterfereValues(a, b *Value) {
	a.interfere[b] = struct{}{}
	b.interfere[a] = struct{}{}
}

func livenessAnalysis(block *Block) {
	liveIn := map[*Value]struct{}{}

	for val := range block.liveOut {
		liveIn[val] = struct{}{}
	}

	switch branch := block.branch.(type) {
	case *ConditionalJump:
		ReviveValue(liveIn, branch.a)
		ReviveValue(liveIn, branch.b)

	case *Exit:
		ReviveValue(liveIn, branch.val)
	}

	for i := range block.instructions {
		inst := block.instructions[len(block.instructions)-i-1]
		switch inst := inst.(type) {
		case *Constant:
			KillValue(liveIn, inst.dest)

		case *Binary:
			KillValue(liveIn, inst.dest)
			ReviveValue(liveIn, inst.a)
			ReviveValue(liveIn, inst.b)

		case *Copy:
			KillValue(liveIn, inst.dest)
			ReviveValue(liveIn, inst.src)
		}
	}

	for _, previous := range block.previousBlocks {
		updated := false
		for val := range liveIn {
			if _, prs := previous.liveOut[val]; !prs {
				previous.liveOut[val] = struct{}{}
				updated = true
			}
		}
		if updated {
			livenessAnalysis(previous)
		}
	}
}
