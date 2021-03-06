package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func runScript(commands []string) []string {
	cmd := exec.Command("go", "run", "main.go", "test.db")
	stdin, _ := cmd.StdinPipe()
	defer stdin.Close()

	for _, command := range commands {
		io.WriteString(stdin, command+"\n")
	}

	// Read entire output
	rawOutput, _ := cmd.CombinedOutput()
	return strings.Split(string(rawOutput), "\n")
}

func TestInsertAndRetrievesARow(t *testing.T) {
	result := runScript([]string{
		"insert 1 user1 person1@example.com",
		"select",
		".exit",
	})
	expect := []string{
		"db > Executed.",
		"db > (1, user1, person1@example.com)",
		"Executed.",
		"db > ",
	}
	defer os.Remove("test.db")

	if !reflect.DeepEqual(result, expect) {
		t.FailNow()
	}
}

func TestTableIsFull(t *testing.T) {
	scripts := make([]string, 302)
	for i := 0; i <= 300; i++ {
		scripts[i] = fmt.Sprintf("insert %d user%d person%d@example.com", i, i, i)
	}
	scripts[len(scripts)-1] = ".exit"
	defer os.Remove("test.db")

	result := runScript(scripts)
	if result[len(scripts)-2] != "db > Error: Table full." {
		t.FailNow()
	}
}

func TestInsertingMaximumLength(t *testing.T) {
	longUsername := strings.Repeat("a", 32)
	longEmail := strings.Repeat("a", 255)
	script := []string{
		fmt.Sprintf("insert 1 %s %s", longUsername, longEmail),
		"select",
		".exit",
	}
	defer os.Remove("test.db")

	result := runScript(script)
	expect := []string{
		"db > Executed.",
		fmt.Sprintf("db > (1, %s, %s)", longUsername, longEmail),
		"Executed.",
		"db > ",
	}
	if !reflect.DeepEqual(result, expect) {
		t.FailNow()
	}
}

func TestStringsAreTooLong(t *testing.T) {
	longUsername := strings.Repeat("a", 33)
	longEmail := strings.Repeat("a", 256)
	script := []string{
		fmt.Sprintf("insert 1 %s %s", longUsername, longEmail),
		"select",
		".exit",
	}
	defer os.Remove("test.db")

	result := runScript(script)
	expect := []string{
		"db > String is too long.",
		"db > Executed.",
		"db > ",
	}
	if !reflect.DeepEqual(result, expect) {
		t.FailNow()
	}
}

func TestIdIsNegative(t *testing.T) {
	script := []string{
		fmt.Sprintf("insert -1 cstack foo@bar.com"),
		"select",
		".exit",
	}
	defer os.Remove("test.db")

	result := runScript(script)
	expect := []string{
		"db > ID must be positive.",
		"db > Executed.",
		"db > ",
	}
	if !reflect.DeepEqual(result, expect) {
		t.FailNow()
	}
}

func TestDataAfterClosingConnection(t *testing.T) {
	result1 := runScript([]string{
		"insert 1 user1 person1@example.com",
		".exit",
	})
	defer os.Remove("test.db")

	expect1 := []string{
		"db > Executed.",
		"db > ",
	}
	if !reflect.DeepEqual(result1, expect1) {
		t.FailNow()
	}

	result2 := runScript([]string{
		"select",
		".exit",
	})
	expect2 := []string{
		"db > (1, user1, person1@example.com)",
		"Executed.",
		"db > ",
	}
	if !reflect.DeepEqual(result2, expect2) {
		t.FailNow()
	}
}
