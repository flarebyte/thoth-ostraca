package inputpipeline

import "testing"

func TestA(t *testing.T) {
	if A() != "a" {
		t.Fatal("unexpected value")
	}
}
