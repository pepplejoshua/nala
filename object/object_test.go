package object

import "testing"

func TestStringHashKey(t *testing.T) {
	h1 := &String{Value: "Hello World"}
	h2 := &String{Value: "Hello World"}
	h3 := &String{Value: "My name is Joshua"}
	h4 := &String{Value: "My name is Joshua"}

	if h1.HashKey() != h2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}

	if h3.HashKey() != h4.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}

	if h1.HashKey() == h3.HashKey() {
		t.Errorf("strings with different content have same hash keys")
	}
}
