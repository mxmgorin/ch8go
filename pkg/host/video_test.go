package host

import "testing"

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		in   string
		want Color
	}{
		{"#aabbcc", Color{0xaa, 0xbb, 0xcc, 255}},
		{"aabbcc", Color{0xaa, 0xbb, 0xcc, 255}}, // "#" prefix is optional
		{"#FF0000", Color{255, 0, 0, 255}},
		{"#000000", Color{0, 0, 0, 255}},
	}
	for _, tt := range tests {
		got, err := ParseHexColor(tt.in)
		if err != nil {
			t.Errorf("ParseHexColor(%q) unexpected error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseHexColor(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestParseHexColorErrors(t *testing.T) {
	for _, in := range []string{"", "12345", "1234567", "#gggggg"} {
		if _, err := ParseHexColor(in); err == nil {
			t.Errorf("ParseHexColor(%q) expected an error, got nil", in)
		}
	}
}

func TestColorToHex(t *testing.T) {
	if got := (Color{0xaa, 0xbb, 0xcc, 255}).ToHex(); got != "#aabbcc" {
		t.Errorf("ToHex() = %q, want %q", got, "#aabbcc")
	}

	// Round-trip: parse then format yields the original (lowercased) hex.
	const hex = "#123abc"
	c, err := ParseHexColor(hex)
	if err != nil {
		t.Fatalf("ParseHexColor(%q) error = %v", hex, err)
	}
	if got := c.ToHex(); got != hex {
		t.Errorf("round-trip ToHex() = %q, want %q", got, hex)
	}
}

func TestPaletteSetColor(t *testing.T) {
	var p Palette

	if err := p.SetColor(0, "#ff0000"); err != nil {
		t.Fatalf("SetColor valid hex error = %v", err)
	}
	if p.Pixels[0] != (Color{255, 0, 0, 255}) {
		t.Errorf("Pixels[0] = %v, want {255 0 0 255}", p.Pixels[0])
	}

	if err := p.SetColor(1, "not-a-color"); err == nil {
		t.Error("SetColor with invalid hex should return an error")
	}
}

func TestNewPalette(t *testing.T) {
	p, err := NewPalette([]string{"#000000", "#ffffff"}, "#ff0000", "#00ff00")
	if err != nil {
		t.Fatalf("NewPalette error = %v", err)
	}
	if p.Pixels[0] != (Color{0, 0, 0, 255}) {
		t.Errorf("Pixels[0] = %v, want black", p.Pixels[0])
	}
	if p.Pixels[1] != (Color{255, 255, 255, 255}) {
		t.Errorf("Pixels[1] = %v, want white", p.Pixels[1])
	}
	if p.Buzzer != (Color{255, 0, 0, 255}) {
		t.Errorf("Buzzer = %v, want red", p.Buzzer)
	}
	if p.Silence != (Color{0, 255, 0, 255}) {
		t.Errorf("Silence = %v, want green", p.Silence)
	}

	if _, err := NewPalette(nil, "zzzzzz", ""); err == nil {
		t.Error("NewPalette with invalid buzzer color should return an error")
	}
}
