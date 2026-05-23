package ocrapi

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetTextFromFileExtractablePDFSkipsOCR(t *testing.T) {
	path := writeTextPDF(t, "Extractable PDF text 123")
	service := NewService("http://127.0.0.1:1", "")

	text, err := service.GetTextFromFile(path)
	if err != nil {
		t.Fatalf("GetTextFromFile() error = %v", err)
	}

	if !strings.Contains(text, "Extractable PDF text 123") {
		t.Fatalf("GetTextFromFile() = %q, want extracted PDF text", text)
	}
}

func TestHasUsableTextHandlesCyrillic(t *testing.T) {
	if !hasUsableText("Текст") {
		t.Fatal("hasUsableText() should treat Cyrillic letters as usable text")
	}
}

func writeTextPDF(t *testing.T, text string) string {
	t.Helper()

	var buf bytes.Buffer
	offsets := []int{0}
	writeObj := func(id int, body string) {
		offsets = append(offsets, buf.Len())
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", id, body)
	}

	buf.WriteString("%PDF-1.4\n")
	writeObj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	writeObj(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	writeObj(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>")

	stream := fmt.Sprintf("BT /F1 24 Tf 72 720 Td (%s) Tj ET", escapePDFString(text))
	writeObj(4, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream))
	writeObj(5, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")

	xrefOffset := buf.Len()
	buf.WriteString("xref\n0 6\n0000000000 65535 f \n")
	for i := 1; i <= 5; i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefOffset)

	path := filepath.Join(t.TempDir(), "extractable.pdf")
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}

func escapePDFString(text string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `(`, `\(`, `)`, `\)`)
	return replacer.Replace(text)
}
