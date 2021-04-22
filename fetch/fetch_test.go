package fetch

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func Test_findFile(t *testing.T) {
	type args struct {
		list    []string
		pattern string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "should find item",
			args:    args{[]string{"aaa", "aaa bbb CCC ddd"}, "aaa bbb CCC ddd"},
			want:    "aaa bbb CCC ddd",
			wantErr: false,
		},
		{
			name:    "should not find item",
			args:    args{[]string{"aaa", "aaa bbb CCC ddd"}, "aaa bbb xCC ddd"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findFile(tt.args.list, tt.args.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("findFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
