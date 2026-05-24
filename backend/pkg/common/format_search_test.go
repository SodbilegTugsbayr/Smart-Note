package common

import "testing"

func TestCleanStringRemovesNullBytesAndTrims(t *testing.T) {
	got := CleanString(" \x00Smart Note\x00 \n")
	if got != "Smart Note" {
		t.Fatalf("CleanString() = %q, want %q", got, "Smart Note")
	}
}

func TestFormatAmountUsesMongolianDisplaySeparators(t *testing.T) {
	got := FormatAmount(1234.5)
	want := "1 234,50"
	if got != want {
		t.Fatalf("FormatAmount() = %q, want %q", got, want)
	}
}

func TestFindReturnsIndexAndFoundFlag(t *testing.T) {
	index, ok := Find([]string{"course", "note", "quiz"}, "note")
	if !ok || index != 1 {
		t.Fatalf("Find() = (%d, %v), want (1, true)", index, ok)
	}

	index, ok = Find([]string{"course", "note", "quiz"}, "missing")
	if ok || index != -1 {
		t.Fatalf("Find() missing = (%d, %v), want (-1, false)", index, ok)
	}
}

func TestFindIntReturnsIndexAndFoundFlag(t *testing.T) {
	index, ok := FindInt([]int{10, 20, 30}, 30)
	if !ok || index != 2 {
		t.Fatalf("FindInt() = (%d, %v), want (2, true)", index, ok)
	}

	index, ok = FindInt([]int{10, 20, 30}, 99)
	if ok || index != -1 {
		t.Fatalf("FindInt() missing = (%d, %v), want (-1, false)", index, ok)
	}
}
