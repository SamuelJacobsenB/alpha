package ir

import "fmt"

func DumpModule(m *Module) {
	fmt.Println("Module:")
	for _, f := range m.Functions {
		fmt.Printf("Function %s:\n", f.Name)
		for _, b := range f.Blocks {
			fmt.Printf("  Block %s:\n", b.Name)
			for _, ins := range b.Instr {
				fmt.Printf("    %s\n", ins)
			}
		}
	}
}
