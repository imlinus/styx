
## Examples

### Hello World

```lua
print('Hello World')
```

### Variables

```lua
let name = 'Linus'
let age = 33

print(`Name: ${name}`)
print(`Age: ${age}`)
```

### Strings

```lua
// Single or double quotes with escape sequences
let greeting = 'Hello\tWorld\n'
let path = "C:\\Users\\linus"

// Backtick template literals (multiline + interpolation)
let name = 'Linus'
let html = `
    <h1>Hello ${name}!</h1>
    <p>Welcome back.</p>
`
```

### Functions

```lua
function greet (name)
	print(`Hello ${name}`)
end

greet('Linus')
```

### Loops

```lua
let i = 0
loop i < 5
	i = i + 1
	print(i)
end
```

### Conditionals

```lua
if age >= 18 then
	print('Adult')
else
	print('Minor')
end
```

### Arrays

```lua
let colors = ['red', 'blue', 'yellow']

loop i, color in colors
	print(color)
end
```

### Objects / Dictionaries

```lua
let person = {
	name: 'Linus',
	age: 33
}

print(person.name)
print(person.age)
```

### Modules

```lua
import hello from './hello.sx'

print(hello('Linus')) // Hello, Linus!
```

### Web Server (Stateless)

Styx uses a file-based routing system (PHP-style). Each `.sx` file is an entry point.

```lua
// api.sx
let name = http.get.name or 'Guest'

http.header('Content-Type', 'application/json')

print({
	message: `Hello ${name}`,
	time: time.iso()
})
```

