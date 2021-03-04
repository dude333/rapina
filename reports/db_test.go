package reports

import "testing"

func Test_avg(t *testing.T) {
	type args struct {
		nums []float32
	}
	tests := []struct {
		name string
		args args
		want float32
	}{
		{"average 2", args{[]float32{1, 2, 3}}, 2},
		{"average 10", args{[]float32{10, 2, 18}}, 10},
		{"average 12.705", args{[]float32{6, 20.4, 18.1, 6.32}}, 12.705},
		{"average 5.5", args{[]float32{5.5, 0}}, 5.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := avg(tt.args.nums...); got != tt.want {
				t.Errorf("avg() = %v, want %v", got, tt.want)
			}
		})
	}
}
