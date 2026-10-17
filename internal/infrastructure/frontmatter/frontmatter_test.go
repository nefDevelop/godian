package frontmatter

import "testing"

func TestSplit(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		wantFM   string
		wantBody string
		wantErr  bool
	}{
		{
			name:     "con frontmatter",
			in:       "---\ntitle: Hola\n---\ncuerpo\n",
			wantFM:   "title: Hola\n",
			wantBody: "cuerpo\n",
		},
		{
			name:     "sin frontmatter",
			in:       "# Titulo\n",
			wantBody: "# Titulo\n",
		},
		{
			name:     "finales de linea CRLF",
			in:       "---\r\ntitle: Hola\r\n---\r\ncuerpo\r\n",
			wantFM:   "title: Hola\n",
			wantBody: "cuerpo\r\n",
		},
		{
			name:     "sin delimitador de cierre",
			in:       "---\ntitle: Hola\ncuerpo\n",
			wantBody: "---\ntitle: Hola\ncuerpo\n",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := Split([]byte(tt.in))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Split() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if string(fm) != tt.wantFM {
				t.Errorf("Split() fm = %q, want %q", string(fm), tt.wantFM)
			}
			if string(body) != tt.wantBody {
				t.Errorf("Split() body = %q, want %q", string(body), tt.wantBody)
			}
		})
	}
}

func FuzzSplit(f *testing.F) {
	f.Add([]byte("---\ntitle: x\n---\nbody"))
	f.Add([]byte("---\r\nk: v\r\n---\r\ncuerpo"))
	f.Add([]byte("sin frontmatter"))
	f.Add([]byte("---\nsin cierre\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		fm, body, _ := Split(data)
		// Invariante: la suma no debe superar el original de forma absurda.
		if len(fm)+len(body) > len(data)+len(data) {
			t.Fatalf("resultado anómalo: fm=%d body=%d in=%d", len(fm), len(body), len(data))
		}
	})
}

func TestUnmarshal(t *testing.T) {
	type doc struct {
		Title string   `yaml:"title"`
		Tags  []string `yaml:"tags"`
	}

	var d doc
	if err := Unmarshal([]byte("title: Hola\ntags: [a, b]\n"), &d); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if d.Title != "Hola" {
		t.Errorf("Title = %q, want %q", d.Title, "Hola")
	}
	if len(d.Tags) != 2 {
		t.Errorf("Tags = %v, want 2 elementos", d.Tags)
	}

	var empty doc
	if err := Unmarshal(nil, &empty); err != nil {
		t.Fatalf("Unmarshal(nil) error = %v", err)
	}
}
