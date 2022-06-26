package generator

import (
	"fmt"
	"testing"
)

func TestGenerator(t *testing.T) {
	fmt.Println("generator")
	generator := CreateGenerator()
	fmt.Println(generator)
}
