package burrow

import (
	"testing"
)

func TestMain(t *testing.T) {
	b := Directory{
		Path:     "/home/alireza/GoPrj/Clonopher/tmp",
		Children: make(map[string]*Directory),
	}
	err := b.Explore()
	if err != nil {
		t.Log(err)
	}
	// fmt.Println(b)
}
