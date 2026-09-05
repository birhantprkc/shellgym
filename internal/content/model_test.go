package content

import "testing"

func TestSplitPrefix(t *testing.T) {
	tests := []struct {
		folder    string
		wantOrder int
		wantName  string
		wantErr   bool
	}{
		{folder: "010.warm-up", wantOrder: 10, wantName: "warm-up"},
		{folder: "020.foo.bar", wantErr: true},
		{folder: "030.Foo", wantErr: true},
		{folder: "040.-x", wantErr: true},
		{folder: "abc", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.folder, func(t *testing.T) {
			order, name, err := splitPrefix(tt.folder)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("splitPrefix(%q) = %d, %q, <nil>; want error", tt.folder, order, name)
				}
				return
			}
			if err != nil {
				t.Fatalf("splitPrefix(%q) unexpected error: %v", tt.folder, err)
			}
			if order != tt.wantOrder || name != tt.wantName {
				t.Fatalf("splitPrefix(%q) = %d, %q; want %d, %q", tt.folder, order, name, tt.wantOrder, tt.wantName)
			}
		})
	}
}
