package tui

import "testing"

func TestSanitizeProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "simple", input: "my-project", want: "my-project"},
		{name: "spaces allowed", input: "My New App", want: "My New App"},
		{name: "trim outer spaces", input: "  test  ", want: "test"},
		{name: "empty", input: "", wantErr: true},
		{name: "slash", input: "a/b", wantErr: true},
		{name: "backslash", input: "a\\b", wantErr: true},
		{name: "dot", input: ".", wantErr: true},
		{name: "dotdot", input: "..", wantErr: true},
		{name: "reserved", input: "CON", wantErr: true},
		{name: "invalid chars", input: "a:b", wantErr: true},
		{name: "trailing dot", input: "abc.", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitizeProjectName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("sanitizeProjectName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
