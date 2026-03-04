package threshold

import "testing"

// --- Config.Validate ---

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		// Valid configs.
		{name: "2-of-3 sss", config: Config{2, 3, SSS}},
		{name: "2-of-2 feldman", config: Config{2, 2, Feldman}},
		{name: "3-of-5 pedersen", config: Config{3, 5, Pedersen}},

		// k < 2.
		{name: "k=0", config: Config{0, 3, SSS}, wantErr: true},
		{name: "k=1", config: Config{1, 3, SSS}, wantErr: true},
		{name: "k=-1", config: Config{-1, 3, SSS}, wantErr: true},

		// k > n.
		{name: "k>n", config: Config{4, 3, SSS}, wantErr: true},

		// Invalid scheme.
		{name: "empty scheme", config: Config{2, 3, ""}, wantErr: true},
		{name: "unknown scheme", config: Config{2, 3, "rsa"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// --- Scheme.Valid ---

func TestSchemeValid(t *testing.T) {
	for _, s := range ValidSchemes {
		if !s.Valid() {
			t.Errorf("scheme %q should be valid", s)
		}
	}
	for _, s := range []Scheme{"", "rsa", "bls", "Feldman"} {
		if s.Valid() {
			t.Errorf("scheme %q should be invalid", s)
		}
	}
}

// --- Phase transitions ---

func TestPhaseTransitionsLegal(t *testing.T) {
	legal := []struct {
		from, to Phase
	}{
		{Pending, Splitting},
		{Splitting, Active},
		{Active, Refreshing},
		{Active, Degraded},
		{Refreshing, Active},
		{Refreshing, Failed},
	}
	for _, tt := range legal {
		if !tt.from.CanTransition(tt.to) {
			t.Errorf("%s → %s should be legal", tt.from, tt.to)
		}
	}
}

func TestPhaseTransitionsIllegal(t *testing.T) {
	illegal := []struct {
		from, to Phase
	}{
		// No backward transitions.
		{Splitting, Pending},
		{Active, Splitting},
		{Active, Pending},
		{Refreshing, Splitting},

		// Terminal states have no outgoing edges.
		{Degraded, Active},
		{Degraded, Pending},
		{Failed, Active},
		{Failed, Refreshing},

		// Self-transitions are not allowed.
		{Pending, Pending},
		{Active, Active},
		{Refreshing, Refreshing},

		// Skip transitions.
		{Pending, Active},
		{Splitting, Refreshing},
	}
	for _, tt := range illegal {
		if tt.from.CanTransition(tt.to) {
			t.Errorf("%s → %s should be illegal", tt.from, tt.to)
		}
	}
}

func TestTerminalPhasesHaveNoTransitions(t *testing.T) {
	for _, p := range []Phase{Degraded, Failed} {
		all := []Phase{Pending, Splitting, Active, Refreshing, Degraded, Failed}
		for _, next := range all {
			if p.CanTransition(next) {
				t.Errorf("terminal phase %s should not transition to %s", p, next)
			}
		}
	}
}

// --- Holder uniqueness ---

func TestHoldersUnique(t *testing.T) {
	tests := []struct {
		name    string
		holders []Holder
		want    bool
	}{
		{name: "empty", holders: nil, want: true},
		{name: "single", holders: []Holder{{ID: "a"}}, want: true},
		{name: "distinct", holders: []Holder{{ID: "a"}, {ID: "b"}, {ID: "c"}}, want: true},
		{name: "duplicate", holders: []Holder{{ID: "a"}, {ID: "b"}, {ID: "a"}}, want: false},
		{name: "all same", holders: []Holder{{ID: "x"}, {ID: "x"}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HoldersUnique(tt.holders); got != tt.want {
				t.Errorf("HoldersUnique() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHoldersReadyIndependentOfUniqueness(t *testing.T) {
	// Ready is orthogonal to uniqueness — different Ready values
	// on holders with the same ID are still duplicates.
	holders := []Holder{
		{ID: "a", Ready: true},
		{ID: "a", Ready: false},
	}
	if HoldersUnique(holders) {
		t.Error("same ID with different Ready should still be duplicate")
	}
}
