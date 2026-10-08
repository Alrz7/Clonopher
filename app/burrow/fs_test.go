package burrow

import (
	"fmt"
	"testing"
)

func TestMain(t *testing.T) {

	b := Stage{
		Path:  "/home/alireza/GoPrj/Clonopher/tmp",
		Roots: make(map[string]*Stage),
	}
	err := b.Explore()
	if err != nil {
		t.Log(err)
	}
	fmt.Println(b)
}
