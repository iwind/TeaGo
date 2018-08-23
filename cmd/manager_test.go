package cmd

import "testing"

func TestParseArgs(t *testing.T) {
	for _, arg := range ParseArgs("run arg0 \"arg 1\" 'arg 2' '\"arg 3' 'this is \\' arg4 \\ ' a") {
		t.Log("Arg:", arg)
	}
}
