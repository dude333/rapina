package reports

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf)

	tests := []struct {
		name string
		fn   func(format string, v ...interface{})
		msg  string
		want string
	}{
		{
			name: "printf",
			fn:   log.Printf,
			msg:  "this is a normal message",
			want: "this is a normal message",
		},
		{
			name: "trace",
			fn:   log.Trace,
			msg:  "this is a trace log\n",
			want: "[TRACE] this is a trace log\n",
		},
		{
			name: "debug",
			fn:   log.Debug,
			msg:  "this is a debug log",
			want: "[DEBUG] this is a debug log\n",
		},
		{
			name: "info",
			fn:   log.Info,
			msg:  "you have been informed",
			want: "[INFO] you have been informed\n",
		},
		{
			name: "warn",
			fn:   log.Warn,
			msg:  "this is a warning",
			want: "[WARN] this is a warning\n",
		},
		{
			name: "error",
			fn:   log.Error,
			msg:  "this is an error message",
			want: "[ERROR] this is an error message\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(tt.msg)
			assert.Equal(t, tt.want, buf.String())
			buf.Reset()
		})
	}
}

func TestLogger_Run(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf)

	log.Run("starting process %d\n", 10)
	log.Ok()
	assert.Equal(t, "[ ] starting process 10\r[✓]\n", buf.String())
	buf.Reset()

	log.Run("starting process %d", 15)
	log.Nok()
	assert.Equal(t, "[ ] starting process 15\r[✗]\n", buf.String())
}
