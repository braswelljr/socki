package condition

import (
	"fmt"
	"testing"
)

// TestTernary - test the Ternary function
//
//	@param t - testing.T
func TestTernary(t *testing.T) {
	// create a slice
	slice := []struct {
		condition bool
		a         interface{}
		b         interface{}
		result    interface{}
	}{
		{
			condition: true,
			a:         "a",
			b:         "b",
			result:    "a",
		},
		{
			condition: false,
			a:         "a",
			b:         "b",
			result:    "b",
		},
	}

	// call the function
	// check if the slice contains the value
	for _, item := range slice {
		t.Run(fmt.Sprintf("test %s", item.result), func(t *testing.T) {

			// check if the condition is true
			if Ternary(item.condition, item.a, item.b) != item.result {
				// log the error
				t.Errorf("condition %v should return %v", item.condition, item.result)
			}
		})
	}
}
