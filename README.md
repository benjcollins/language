# General Purpose Programming Language for Embedded Systems

The project can be build with `go build` or `go run main.go` on linux systems.
This will compile the code in the file example.txt.
It will produce 4 output files all located in the `build` folder:
- `output.ast` - A textual representation of the Abstract Syntax Tree for the source program.
- `output.ir` - A textual representation of the intermediate representation of the source program. This is basically an RTL (Register Transfer Language).
- `output.s` - The source program converted to optimised x86 assembly.
- `output` - This is the executable built using gcc without a standard libary to create small binaries. This binary will only run on linux systems because it uses Sys calls rather than the Win32 API because they are simpler.

To run the output file simply type `./build/output` to execute it. The output of the file is stored in the exit code which can be accessed by executing the command `echo $?`. Take note however that due to backwards compatibility this is only an 8 bit number so numbers greater than 255 will overflow.
