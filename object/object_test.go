package object

import (
	"fmt"
	"testing"
)

func TestStringHashKey(t *testing.T) {
	hello1 := &String{Value: "Hello World"}
	hello2 := &String{Value: "Hello World"}
	diff1 := &String{Value: "My name is johnny"}
	diff2 := &String{Value: "My name is johnny"}

	if hello1.HashKey() != hello2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}

	if diff1.HashKey() != diff2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}

	if hello1.HashKey() == diff1.HashKey() {
		t.Errorf("strings with different content have same hash keys")
	}
}

func TestIntBoolHashKey(t *testing.T) {
	oneInt := &Integer{Value: 1}
	trueBool := &Boolean{Value: true}

	zeroInt := &Integer{Value: 0}
	falseBool := &Boolean{Value: false}

	fmt.Println(oneInt.HashKey(), trueBool.HashKey())

	if oneInt.HashKey() == trueBool.HashKey() {
		t.Errorf("int and bool have same hash keys")
	}

	if zeroInt.HashKey() == falseBool.HashKey() {
		t.Errorf("int and bool have same hash keys")
	}
}
