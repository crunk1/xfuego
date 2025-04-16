package types

import (
	"reflect"
	"testing"
)

func TestIsNullable(t *testing.T) {
	tests := []struct {
		name string
		t    reflect.Type
		want bool
	}{
		{"F:IsNullable(<int>)", reflect.TypeOf(0), false},
		{"F:IsNullable(<*int>)", reflect.TypeOf((*int)(nil)), false},
		{"F:IsNullable(<any>)", reflect.TypeOf((*any)(nil)), false},
		{"T:IsNullable(<Nullable[int]>)", reflect.TypeOf((*Nullable[int])(nil)).Elem(), true},
		{"F:IsNullable(<*Nullable[int]>)", reflect.TypeOf((*Nullable[int])(nil)), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNullable(tt.t); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
