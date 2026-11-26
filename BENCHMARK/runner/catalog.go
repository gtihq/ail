package runner

import "fmt"

const goMod = "module example.com/bench\n\ngo 1.25\n"

const mainTest = `package main

import "testing"

func TestAdd(t *testing.T) {
	if add(2, 3) != 5 {
		t.Fatalf("add(2, 3) = %d, want 5", add(2, 3))
	}
}
`

func mainSource(buggy bool) string {
	body := "\treturn a + b"
	if buggy {
		body = "\treturn a - b"
	}
	return fmt.Sprintf(`package main

func add(a, b int) int {
%s
}

func main() {}
`, body)
}

type categorySpec struct {
	prefix string
	name   string
	count  int
	rules  []string
	prompt func(int) string
	buggy  bool
}

func implementationPrompt(index int) string {
	topics := []string{
		"return the larger of two integers",
		"return the absolute value of an integer",
		"clamp a value between a lower and upper bound",
		"count the even numbers in a slice",
		"reverse a string without changing its length",
		"compute the factorial of a small non-negative integer",
		"return the index of a value in a slice or minus one",
		"merge two sorted slices into one sorted slice",
		"check whether a string is a palindrome",
		"sum all integers in a slice",
	}
	return fmt.Sprintf("Implement a helper in main.go that %s. Keep the existing add function working. Run go test ./... when done.", topics[index%len(topics)])
}

func debuggingPrompt(index int) string {
	targets := []string{"add", "the arithmetic helper", "the sum calculation", "the result computation", "the integer operation", "the core function", "the value returned by add", "the addition logic", "the numeric helper", "the small calculation"}
	return fmt.Sprintf("The %s in main.go is incorrect and the test fails. Fix the bug without rewriting unrelated code. Run go test ./... to verify.", targets[index%len(targets)])
}

func refactoringPrompt(index int) string {
	goals := []string{
		"extract the addition into a named helper with a clear comment",
		"replace the magic numbers in the test with named constants",
		"split addition and validation into separate functions",
		"introduce a small type to represent a pair of integers",
		"rename identifiers to follow Go conventions",
		"remove duplication between add and the test setup",
		"use a table-driven structure in the test",
		"simplify main without changing behavior",
		"group related declarations together",
		"extract a reusable check function",
	}
	return fmt.Sprintf("Refactor main.go to %s. Behavior must not change. Run go test ./... when done.", goals[index%len(goals)])
}

func testingPrompt(index int) string {
	targets := []string{"add with negative numbers", "add with zero", "add with large values", "the main entry point", "add with mixed signs"}
	return fmt.Sprintf("Add table-driven tests in main_test.go covering %s. Do not change main.go. Run go test ./... when done.", targets[index%len(targets)])
}

func architecturePrompt(index int) string {
	goals := []string{
		"introduce a small package boundary for arithmetic helpers",
		"separate computation from the main entry point",
		"add a narrow interface for integer operations",
		"group helpers into a focused file without exceeding size limits",
		"design a minimal structure that makes the helper easy to test",
	}
	return fmt.Sprintf("Propose and implement a small architecture change in main.go: %s. Keep the module building and tests passing.", goals[index%len(goals)])
}

func reviewPrompt(index int) string {
	focus := []string{"correctness", "error handling", "naming", "simplicity", "test quality"}
	return fmt.Sprintf("Review main.go and main_test.go focusing on %s. List concrete issues. Do not modify any files.", focus[index%len(focus)])
}

func ambiguousPrompt(index int) string {
	hints := []string{
		"Make this code better.",
		"Improve the small program.",
		"Clean up the implementation.",
		"Tidy the project a bit.",
		"Can you make this nicer?",
	}
	return hints[index%len(hints)]
}

// GenerateCatalog builds the 50-task benchmark catalogue deterministically.
func GenerateCatalog() Catalog {
	specs := []categorySpec{
		{
			prefix: "IMP", name: "implementation", count: 10,
			rules:  []string{"R001", "R002", "R003", "R005", "R006"},
			prompt: implementationPrompt,
		},
		{
			prefix: "DBG", name: "debugging", count: 10,
			rules:  []string{"R003", "R005", "R006"},
			prompt: debuggingPrompt, buggy: true,
		},
		{
			prefix: "REF", name: "refactoring", count: 10,
			rules:  []string{"R001", "R002", "R005", "R007"},
			prompt: refactoringPrompt,
		},
		{
			prefix: "TST", name: "testing", count: 5,
			rules:  []string{"R003", "R006"},
			prompt: testingPrompt,
		},
		{
			prefix: "ARC", name: "architecture", count: 5,
			rules:  []string{"R002", "R005", "R007"},
			prompt: architecturePrompt,
		},
		{
			prefix: "REV", name: "code-review", count: 5,
			rules:  []string{"R006", "R010"},
			prompt: reviewPrompt,
		},
		{
			prefix: "AMB", name: "ambiguous-requirement", count: 5,
			rules:  []string{"R005", "R006", "R008"},
			prompt: ambiguousPrompt,
		},
	}

	tasks := make([]Task, 0, 50)
	for _, spec := range specs {
		for index := 0; index < spec.count; index++ {
			id := fmt.Sprintf("%s-%02d", spec.prefix, index+1)
			files := map[string]string{
				"go.mod":      goMod,
				"main.go":     mainSource(spec.buggy),
				"main_test.go": mainTest,
			}
			tasks = append(tasks, Task{
				ID:              id,
				Category:        spec.name,
				Prompt:          spec.prompt(index),
				Files:           files,
				ApplicableRules: spec.rules,
				VerifyCommand:   "go test ./...",
			})
		}
	}
	return Catalog{Version: "0.3", Tasks: tasks}
}
