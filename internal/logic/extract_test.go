package logic

import "testing"

func TestExtractDate(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "slash separated wiki cell",
			raw:  "Update / Jun 18, 2026",
			want: "2026-06-18",
		},
		{
			name: "long month slash separated wiki cell",
			raw:  "June 4, 2026 / June 4, 2026",
			want: "2026-06-04",
		},
		{
			name: "date inside row text",
			raw:  "Update Date Jun 18, 2026",
			want: "2026-06-18",
		},
		{
			name: "iso date",
			raw:  "2026-06-18",
			want: "2026-06-18",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractDate(tt.raw)
			if err != nil {
				t.Fatalf("ExtractDate() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ExtractDate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractDateRejectsEmptyInput(t *testing.T) {
	if _, err := ExtractDate(""); err == nil {
		t.Fatal("ExtractDate() error = nil, want error")
	}
}
