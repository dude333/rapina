package rapina

import (
	"reflect"
	"testing"
	"time"
)

func TestIsDate(t *testing.T) {
	type args struct {
		date string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should be true",
			args: args{date: "2021-04-26"},
			want: true,
		},
		{
			name: "should be true too",
			args: args{date: "2030-12-31"},
			want: true,
		},
		{
			name: "should be false",
			args: args{date: "2021-04-31"},
			want: false,
		},
		{
			name: "should be false too",
			args: args{date: "20/12/2000"},
			want: false,
		},
		{
			name: "should be false three",
			args: args{date: "2021-07-32"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDate(tt.args.date); got != tt.want {
				t.Errorf("IsDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUrl(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should be true",
			args: args{str: "http://example.com/path"},
			want: true,
		},
		{
			name: "should be false",
			args: args{str: "example.com/path"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsURL(tt.args.str); got != tt.want {
				t.Errorf("IsUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMonthsFromToday(t *testing.T) {
	timeNow1 := func() time.Time {
		return time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	}
	timeNow2 := func() time.Time {
		return time.Date(2009, time.March, 31, 23, 0, 0, 0, time.UTC)
	}

	type args struct {
		n int
	}
	tests := []struct {
		name    string
		args    args
		timeNow func() time.Time
		want    []string
	}{
		{
			name:    "should show 3 months",
			args:    args{n: 3},
			timeNow: timeNow1,
			want:    []string{"2009-11", "2009-10", "2009-09"},
		},
		{
			name:    "should show 2 months",
			args:    args{n: 2},
			timeNow: timeNow2,
			want:    []string{"2009-03", "2009-02"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_timeNow = tt.timeNow
			if got := MonthsFromToday(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MonthsFromToday() = %#v, want %v", got, tt.want)
			}
		})
	}
}
