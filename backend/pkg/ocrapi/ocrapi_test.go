package ocrapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestCleanExtractedPDFTextTrimsLinesAndCollapsesBlankRuns(t *testing.T) {
	got := cleanExtractedPDFText("  first line\r\n\r\n\r\n second line \r third line ")
	want := "first line\n\nsecond line\nthird line"
	if got != want {
		t.Fatalf("cleanExtractedPDFText() = %q, want %q", got, want)
	}
}

func TestPushFilePostsSupportedFileToOCRAPI(t *testing.T) {
	path := writeBinaryFile(t, "scan.png", minimalPNGBytes())
	var receivedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); got != "image/png" {
			t.Fatalf("Content-Type = %q, want image/png", got)
		}
		if got := r.Header.Get("Token"); got != "ocr-token" {
			t.Fatalf("Token = %q, want ocr-token", got)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		receivedBody = body

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(OCRResult{FileName: "scan.png", PageCount: 1, UUID: "job-1"}); err != nil {
			t.Fatalf("Encode() error = %v", err)
		}
	}))
	defer server.Close()

	service := NewService(server.URL, "ocr-token")
	result, err := service.pushFile(path)
	if err != nil {
		t.Fatalf("pushFile() error = %v", err)
	}

	if result.UUID != "job-1" || result.PageCount != 1 {
		t.Fatalf("pushFile() result = %+v, want UUID job-1 and one page", result)
	}
	if !bytes.Equal(receivedBody, minimalPNGBytes()) {
		t.Fatal("OCR server received unexpected file body")
	}
}

func TestPushFileRejectsUnsupportedFileType(t *testing.T) {
	path := writeBinaryFile(t, "notes.txt", []byte("plain text"))
	service := NewService("http://127.0.0.1:1", "ocr-token")

	_, err := service.pushFile(path)
	if err == nil {
		t.Fatal("pushFile() error = nil, want unsupported file type")
	}
	if !strings.Contains(err.Error(), "unsupported file type") {
		t.Fatalf("pushFile() error = %q, want unsupported file type", err.Error())
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

func writeBinaryFile(t *testing.T, name string, data []byte) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func minimalPNGBytes() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}
}
