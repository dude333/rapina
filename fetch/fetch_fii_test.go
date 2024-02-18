package fetch

import (
	"testing"
)

func Test_comma2dot(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "should work",
			args: args{val: "1.230,56"},
			want: 1230.56,
		},
		{
			name: "should return 0",
			args: args{val: "shouldbeanum"},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := comma2dot(tt.args.val); got != tt.want {
				t.Errorf("comma2dot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_FixDate(t *testing.T) {
	type args struct {
		date string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should work",
			args: args{date: "01/02/2021"},
			want: "2021-02-01",
		},
		{
			name: "should return the input",
			args: args{date: "wrong/date"},
			want: "wrong/date",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixDate(tt.args.date); got != tt.want {
				t.Errorf("fixDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
