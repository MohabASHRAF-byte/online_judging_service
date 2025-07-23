package main

import "fmt"

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rectangle struct {
	Width, Height float64
}
type Square struct {
	Width, Height float64
}

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

func (s Square) Area() float64      { return s.Width * 1 }
func (s Square) Perimeter() float64 { return 2 * (s.Width + s.Height) }

func main() {
	r := Rectangle{Width: 3, Height: 4}
	s := Square{Width: 5, Height: 5}

	// Slice of Shape, holding both Rectangle and Square
	shapes := []Shape{r, s}

	// Iterate over shapes and call their methods
	for _, shape := range shapes {
		fmt.Printf("Area: %.1f, Perimeter: %.1f\n", shape.Area(), shape.Perimeter())
	}
}

/*log.Println("Starting Simple Judging Service...")

// Initialize store (submissions are already loaded)
store := store.NewSimpleStore()

// Initialize processor
log.Println("Processing submissions...")

// Get next submission
submission := store.GetNext()
outputs, err := processor.RunCppWithTestcases(submission.Code, submission.TestInputs)
if err != nil {
	log.Printf("Error executing C++ code: %v", err)
	return
}
fmt.Println("\n--- Test Results ---")

for i, output := range outputs {
	fmt.Printf("Testcase %d:\n", i+1)
	fmt.Printf("  Input: %s\n", submission.TestInputs[i])
	fmt.Printf("  Output:", output, "\n")
	fmt.Println("  Status: ✓ SUCCESS")
	fmt.Println()
}

fmt.Printf("Total testcases: %d\n", len(outputs))
fmt.Println("=========================")*/
