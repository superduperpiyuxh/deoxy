package example

import "fmt"

// Add returns the sum of two integers.
func Add(a int, b int) int {
	return a + b
}

// helper is an unexported function used internally.
func helper(name string) string {
	return "hello " + name
}

// Person represents a person with name and age.
type Person struct {
	Name string
	Age  int
}

// Greet returns a greeting from the person.
func (p Person) Greet() string {
	return "Hello, my name is " + p.Name
}

// SetAge sets the age of the person.
func (p *Person) SetAge(age int) {
	p.Age = age
}

// Stringer is an interface for types that can describe themselves.
type Stringer interface {
	String() string
}

// MultiplyAndAdd returns both product and sum.
func MultiplyAndAdd(x, y int) (int, int) {
	return x * y, x + y
}

// FirstOrDefault returns the first element or a default value.
func FirstOrDefault[T any](items []T, defaultVal T) T {
	if len(items) > 0 {
		return items[0]
	}
	return defaultVal
}

// DoNothing is a function with no parameters and no return value.
func DoNothing() {
}
