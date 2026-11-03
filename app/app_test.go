package app

import (
	"clonopher/app/burrow"
	"testing"
)

func TestParseRoot(t *testing.T)  {
	err :=ParseRoots("")
	if err !=nil {
	 t.Log(err)
	}
	t.Log(burrow.Globals)
}