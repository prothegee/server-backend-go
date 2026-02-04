package test_unittest

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"showcase-backend-go/pkg"
)

// --------------------------------------------------------- //

// this test is to check original size from library is not change
func TestSizeDefault(t *testing.T) {
	var str string
	const strSize uintptr = 16
	if unsafe.Sizeof(str) != strSize || reflect.TypeOf(str).Size() != strSize {
		t.Error("ERROR: string size is not as expected")
	}

	var i1 int8
	const i1Size uintptr = 1
	if unsafe.Sizeof(i1) != i1Size || reflect.TypeOf(i1).Size() != i1Size {
		t.Error("ERROR: int8 size is not as expected")
	}

	var i2 int16
	const i2Size uintptr = 2
	if unsafe.Sizeof(i2) != i2Size || reflect.TypeOf(i2).Size() != i2Size {
		t.Error("ERROR: in16 size is not as expected")
	}

	var i3 int32
	const i3Size uintptr = 4
	if unsafe.Sizeof(i3) != i3Size || reflect.TypeOf(i3).Size() != i3Size {
		t.Error("ERROR: int32 size is not as expected")
	}

	var i4 int
	const i4Size uintptr = 8
	if unsafe.Sizeof(i4) != i4Size || reflect.TypeOf(i4).Size() != i4Size {
		t.Error("ERROR: int size is not as expected")
	}

	var i5 int64
	const i5Size uintptr = 8
	if unsafe.Sizeof(i5) != i5Size || reflect.TypeOf(i5).Size() != i5Size {
		t.Error("ERROR: int64 size is not as expected")
	}

	var f1 float32
	const f1Size uintptr = 4
	if unsafe.Sizeof(f1) != f1Size || reflect.TypeOf(f1).Size() != f1Size {
		t.Error("ERROR: float32 size is not as expected")
	}

	var f2 float64
	const f2Size uintptr = 8
	if unsafe.Sizeof(f2) != f2Size || reflect.TypeOf(f2).Size() != f2Size {
		t.Error("ERROR: float64 size is not as expected")
	}
}

// checking size of pkh.ConfigServer, not bounded to any restriction
//
// perhaps 512 bytes is crazy enough
func TestSizeConfig(t *testing.T) {
	var cfg pkg.ConfigServer
	const cfgSizeMax uintptr = 512
	if unsafe.Sizeof(cfg) >= cfgSizeMax || reflect.TypeOf(cfg).Size() >= cfgSizeMax {
		fmt.Printf("WARNING: size of 'pkg.ConfigServer' is greater than %d\n", cfgSizeMax)
	}
}

//	this test meant to be check default size of empty struct type and padding
//
// some behaviour came from C and C++ when squezing memory usage
func TestSizeStruct(t *testing.T) {
	type TypeFoo struct{}
	type TypeBar struct {
		String string
		Number int8
	}
	type TypeBaz struct {
		Number int8
		String string
	}

	t1 := TypeFoo{}
	const t1Size uintptr = 0
	if unsafe.Sizeof(t1) != t1Size || reflect.TypeOf(t1).Size() != t1Size {
		t.Error("ERROR: byte size of empty type struct is unepected\n")
	}

	t2 := TypeBar{String: "String", Number: 1}
	const t2Size uintptr = 24
	if unsafe.Sizeof(t2) != t2Size || reflect.TypeOf(t2).Size() != t2Size {
		t.Errorf("ERROR: size of t2 expecting %d\n", t2Size)
	}

	t3 := TypeBar{Number: 1, String: "String"}
	const t3Size uintptr = 24
	if unsafe.Sizeof(t3) != t3Size || reflect.TypeOf(t3).Size() != t3Size {
		t.Errorf("ERROR: size of t3 expecting %d\n", t3Size)
	}
}
