package main

import (
	"fmt"
	"io/ioutil"
	"language/backend"
	"language/frontend"
	"language/syntax"
	"os"
	"os/exec"
)

func main() {

	data, _ := ioutil.ReadFile("example.txt")
	source := string(data)

	ast, errs := syntax.Parse(source)
	if len(errs) != 0 {
		for _, err := range errs {
			fmt.Println(err)
		}
		return
	}

	f, _ := os.Create("build/output.ast")
	f.WriteString(fmt.Sprint(ast))
	f.Close()

	program, entry, exit, ty, errs := frontend.Compile(ast)
	if ty == nil {
		fmt.Println("Compilation unsucessful:")
		for _, err := range errs {
			fmt.Println(err)
		}
		return
	}

	if ty.Type() != "int" {
		fmt.Println("Return type must be integer!")
		return
	}

	exit.Exit(frontend.ToValues(ty)[0])

	backend.MarkUsedValues(program)
	backend.RemoveDeadCode(program)

	backend.LivenessAnalysis(exit, program)
	backend.CoalesceCopies(program)
	backend.CoalesceBinary(program)

	f, _ = os.Create("build/output.ir")
	f.WriteString(backend.IrToStr(program))
	f.Close()

	backend.Execute(entry)

	backend.LivenessAnalysis(exit, program)
	backend.RegisterAllocation(program)

	f, _ = os.Create("build/output.s")
	f.WriteString(backend.X86(program, entry))
	f.Close()

	exec.Command("gcc", "-nostdlib", "build/output.s", "-o", "build/output").Output()
}
