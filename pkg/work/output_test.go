package work

import (
	"context"
	"math/rand/v2"
	"testing"

	"github.com/drrev/telehandler/internal/safe"
	"github.com/drrev/telehandler/tests/utils"
)

func TestOutputReader_Read(t *testing.T) {
	chacha := rand.NewChaCha8([32]byte{})

	buff := func(sz int, closed bool) *safe.Buffer {
		data := make([]byte, sz)
		chacha.Read(data)

		b := safe.NewBuffer(data)
		b.Write(data)
		if closed {
			b.Close()
		}
		return b
	}

	type fields struct {
		seq int64
		off int64
		max int64
		out *safe.Buffer
	}
	tests := []struct {
		name    string
		ctx     context.Context
		fields  fields
		wantN   int
		wantErr func(error) bool
	}{
		{
			name:    "closed: no data",
			ctx:     context.Background(),
			fields:  fields{out: buff(0, true)},
			wantN:   0,
			wantErr: utils.ErrorTextContains(t, "EOF"),
		},
		{
			name:    "closed: backfill",
			ctx:     context.Background(),
			fields:  fields{max: 45, out: buff(45, true)},
			wantN:   45,
			wantErr: utils.NoError(t),
		},
	}

	data := make([]byte, 4096)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OutputReader{
				seq: tt.fields.seq,
				off: tt.fields.off,
				max: tt.fields.max,
				out: tt.fields.out,
			}
			gotN, err := o.Read(tt.ctx, data)
			if !tt.wantErr(err) {
				t.Errorf("OutputReader.Read() error = %v", err)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("OutputReader.Read() = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}
