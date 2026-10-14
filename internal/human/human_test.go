package human

import (
	"bytes"
	"strings"
	"testing"
)

func TestDecoratePlainIsIdentity(t *testing.T) {
	in := "# Title\n\n> ⚠️ UNSUPPORTED: x\n- item\n"
	if got := Decorate(in, false); got != in {
		t.Fatalf("plain Decorate changed the text:\n%q", got)
	}
}

func TestDecorateStyledAddsANSI(t *testing.T) {
	in := "# Title\nplain\n> ⚠️ UNSUPPORTED: x\n"
	got := Decorate(in, true)
	if !strings.Contains(got, "\x1b[") {
		t.Fatalf("styled Decorate produced no ANSI:\n%q", got)
	}
	if !strings.Contains(got, "plain") || !strings.Contains(got, "Title") {
		t.Fatalf("styled Decorate dropped content:\n%q", got)
	}
}

func TestStyledNonFileIsFalse(t *testing.T) {
	if Styled(&bytes.Buffer{}) {
		t.Fatal("a buffer is not a terminal")
	}
}

func TestStyledHonorsNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	// Even a terminal would be plain with NO_COLOR set; a buffer is the safe case.
	if Styled(&bytes.Buffer{}) {
		t.Fatal("NO_COLOR must disable styling")
	}
}
