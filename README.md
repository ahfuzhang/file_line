# file_line
Like `__FILE__/__LINE__` of C: use go generate to get source code line number at compile time.

I used to use this one function to get the line number of the source code:

```go
func SourceCodeLoc(callDepth int) string {
	_, file, line, ok := runtime.Caller(callDepth)
	if !ok {
		return ""
	}
	file = strings.ReplaceAll(file, "\\", "/")
	arr := strings.Split(file, "/")
	if len(arr) > 3 {
		file = strings.Join(arr[len(arr)-3:], "/")
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func example(){
     Mylogger.Infof("[%s]something happens here", SourceCodeLoc(1))
}
```

The function `runtime.Caller()` works at runtime and consumes more resources.
Is it possible to use the macros `__FILE__` and `__LINE__` to get the code line number like in `C` ?
Finally, this can be easily accomplished with golang AST.

# How to use

1. Install:

```
go install github.com/ahfuzhang/file_line@latest
```

2. Write code that uses a place holder like this where a line number is required.
    
```go
func myCode(){
    Mylogger.Infof("%s: something happens here", "[file.go:123]")
}
```

3. Add the go generate directive to the entry point of the program:

```go
//go:generate file_line -src=./

func main() {
	fmt.Println("use a place holder:", "[file.go:123]")
}
```

4. Execute `go generat` before compiling.
* Before compiling, run `go generate` and replace all place holders in the function call arguments with the correct filename and line number.
* You can also run `file_line -src=./` on the command line.

5. Execute `go build`.
