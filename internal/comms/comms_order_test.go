package comms

import "testing"

func TestMessageOrderingTableDriven(t *testing.T) {
	tests := []struct {
		name   string
		msgs   []Response
		expect map[string][]uint64
	}{
		{
			name: "single channel out of order",
			msgs: []Response{
				{ReplyChannelId: "chan", Seq: 3},
				{ReplyChannelId: "chan", Seq: 1},
				{ReplyChannelId: "chan", Seq: 2},
			},
			expect: map[string][]uint64{"chan": {1, 2, 3}},
		},
		{
			name: "multiple channels interleaved",
			msgs: []Response{
				{ReplyChannelId: "chan1", Seq: 2},
				{ReplyChannelId: "chan2", Seq: 1},
				{ReplyChannelId: "chan1", Seq: 1},
				{ReplyChannelId: "chan2", Seq: 2},
			},
			expect: map[string][]uint64{
				"chan1": {1, 2},
				"chan2": {1, 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := MessageHandler{
				ResponseCh: make(chan Response, len(tt.msgs)),
				expected:   make(map[string]uint64),
				pending:    make(map[string]map[uint64]*Response),
			}

			got := make(map[string][]uint64)
			h.sendOverride = func(r *Response) {
				got[r.ReplyChannelId] = append(got[r.ReplyChannelId], r.Seq)
			}

			for i := range tt.msgs {
				h.enqueueAndSend(&tt.msgs[i])
			}

			for ch, want := range tt.expect {
				seqs := got[ch]
				if len(seqs) != len(want) {
					t.Fatalf("channel %s expected %d messages, got %d", ch, len(want), len(seqs))
				}
				for i := range want {
					if seqs[i] != want[i] {
						t.Fatalf("channel %s at %d expected %d, got %d", ch, i, want[i], seqs[i])
					}
				}
			}
		})
	}
}
