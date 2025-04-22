# Technical Application Configuration Overview

## Startup

```bash
go mod tidy
go get github.com/a-h/templ
go get github.com/a-h/templ/runtime
make go-install-air
make go-install-templ
make build
make watch
```

You may need to delete "workstreams.db" if it was created without the latest tables, or use this tool to edit manually:

[VS Code Extension: `SQLite3 Editor`](https://marketplace.visualstudio.com/items/?itemName=yy0931.vscode-sqlite3-editor)

## Notes

A Tour Of Go
https://go.dev/tour/list

### Go Types

- bool
- string
- int int8 int16 int32 int64
- uint uint8 uint16 uint32 uint64 uintptr
- byte // alias for uint8
- rune // alias for int32
- float32 float64
- complex64 complex128

The example shows variables of several types, and also that variable declarations may be "factored" into blocks,
as with import statements.

The int, uint, and uintptr types are usually 32 bits wide on 32-bit systems and 64 bits wide on 64-bit systems.
When you need an integer value you should use int unless you have a specific reason to use a sized or unsigned integer type.

### Pointers

Go has pointers. A pointer holds the memory address of a value.

The type \*T is a pointer to a T value. Its zero value is nil.

`var p *int`
The & operator generates a pointer to its operand.

```go
i := 42
p = &i
```

The \* operator denotes the pointer's underlying value.

```go
fmt.Println(p) // memory address of i ex. 0xc000184040
fmt.Println(*p) // read i through the pointer p
*p = 21 // set i through the pointer p
```

This is known as "dereferencing" or "indirecting".

Unlike C, Go has no pointer arithmetic

### If + Error Handling logic notes

Try to avoid mixing control flow and success logic inside a conditional block. It is against Go's usual clean separation of logic. The lifecycle of a variable may be beyond the error handling for the function call.

Explicit Example:

```go
workstreams, err := s.workstreamDb.GetAllWorkstreams()
if err != nil {
    http.Error(w, "No workstreams", http.StatusInternalServerError)
    return
}
```

You can streamline if/error handling blocks to still be readable and explicit when there is only "error" coming back from the function. The lifecycle of error is only the if block.

Streamlined Example:

```go
if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	log.Fatalf("Error when running server: %s", err)
}
```

**Consideration: Variables inside if blocks are scoped to it, and garbage collected after.**

### Hero Icons / TailwindCSS

[heroicons](https://heroicons.com/)

- FOSS svg code and icons