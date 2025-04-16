package types

import (
	"testing"
)

func TestIsNoneType(t *testing.T) {
	tests := []struct {
		name   string
		result bool
		want   bool
	}{
		{"T:IsNoneType(<None>)", IsNoneType[None](), true},
		{"F:IsNoneType(<int>)", IsNoneType[int](), false},
		{"F:IsNoneType(<any>)", IsNoneType[any](), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result != tt.want {
				t.Errorf("got %v, want %v", tt.result, tt.want)
			}
		})
	}
}
