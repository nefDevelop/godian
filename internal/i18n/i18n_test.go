package i18n

import "testing"

func TestT(t *testing.T) {
	if got := T("es", "nav.home"); got != "Inicio" {
		t.Errorf("T(es, nav.home) = %q, want Inicio", got)
	}
	if got := T("en", "nav.home"); got != "Home" {
		t.Errorf("T(en, nav.home) = %q, want Home", got)
	}
	// idioma desconocido cae a inglés
	if got := T("fr", "nav.home"); got != "Home" {
		t.Errorf("T(fr, nav.home) = %q, want Home (fallback)", got)
	}
	// clave desconocida devuelve la clave
	if got := T("en", "no.such.key"); got != "no.such.key" {
		t.Errorf("T(unknown) = %q, want la clave", got)
	}
}
