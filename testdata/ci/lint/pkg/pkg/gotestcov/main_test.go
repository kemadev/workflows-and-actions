package main

import "testing"

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "test1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}

func Test_f1(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test1",
			args: args{
				s: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f1(tt.args.s)
		})
	}
}

// Oopsy no test coverage for f2 and f3

// func Test_f2(t *testing.T) {
// 	tests := []struct {
// 		name string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			f2()
// 		})
// 	}
// }

// func Test_f3(t *testing.T) {
// 	tests := []struct {
// 		name string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			f3()
// 		})
// 	}
// }
