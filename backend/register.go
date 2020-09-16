package backend

import "sort"

func RegisterAllocation(program *Program) {
	registers := []string{"%eax", "%ecx", "%edx", "%ebx", "%esp", "%ebp", "%esi", "%edi"}

	stack := []*Value{}
	values := make([]*Value, len(program.values))
	copy(values, program.values)

	for len(values) > 0 {

		sort.Slice(values, func(i, j int) bool {
			return len(values[i].interfere) < len(values[j].interfere)
		})

		val := values[0]
		stack = append(stack, val)
		for neigbour := range val.interfere {
			delete(neigbour.interfere, val)
		}
		values = values[1:]
	}

	for len(stack) > 0 {
		val := stack[len(stack)-1]
		for _, reg := range registers {
			val.register = reg
			for neigbour := range val.interfere {
				if neigbour.register == val.register {
					val.register = ""
					break
				}
			}
			if val.register != "" {
				break
			}
		}
		if val.register == "" {
			panic("BAD")
		}
		stack = stack[:len(stack)-1]
	}
}
