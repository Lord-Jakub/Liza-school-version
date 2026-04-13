# Liza – Simple Scripting Language


**Liza** is a simple, statically-typed, interpreted scripting language created as a demonstration project to understand the principles of programming language design and implementation.

## Features

- Static typing (`int`, `float`, `bool`, `string`, `void` + arrays)
- Support for multi-dimensional dynamic arrays
- Functions with parameters and return values
- Control flow: `if`/`else` and `for` loops
- Recursion
- Simple module system using `namespace` and `import`
- Standard libraries written in Liza (`math`, `array`)
- Basic I/O (`print`, `read`)

## Quick Start

### Using the prebuilt Linux binary

Download the binary from release **[v1.0.0-school](https://github.com/Lord-Jakub/Liza-school-version/releases/tag/v1.0.0-school)**.
> Note: prebuild binary is avalible only on linux

```bash
chmod +x liza
./liza examples/fibonacci.li
```
You can use arguments:
- `-help` – show help message
- `-AST` – output abstract syntax tree into a json file
### Building from source
```bash
git clone https://github.com/Lord-Jakub/Liza-school-version.git
cd Liza-school-version

go build -o liza src/main.go

./liza examples/fibonacci.li
```
## Example program
```Liza
namespace main
func fibonacci(int n) int{
    if n <= 1 return n
    int[n+1] arr
    arr[0] = 0
    arr[1] = 1
    for int i = 2; i<=n; i = i+1{
        arr[i] = arr[i-1] + arr[i-2]
      }
    return arr[n]
  }
func main(){
    for true {
        print("Number: ")
        string input = read()
        if input == "exit" {
            exit(0)
        }
        int n = StringToInt(input)
        print(n, "th number of fibonacci sequence is ",fibonacci(n), "\n")
    }
}
```