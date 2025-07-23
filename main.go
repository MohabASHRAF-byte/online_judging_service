package main

import (
	"fmt"
	"judging-service/containers"
)

func main() {
	var poolManger = containers.NewContainersPoolManger(4)
	fmt.Println(poolManger.Limit)
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
