package star

import (
	"testing"
)

func TestIsBug(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			IsBug()
		})
	}
}

func TestRandAtomic(t *testing.T) {
	type args struct {
		max int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{max: 100}, want: 99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RandAtomic(tt.args.max); got >= tt.want {
				t.Errorf("RandAtomic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSerialKey(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
		{name: "test1", want: "12345678190"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(GetSerialKey())
		})
	}
}
