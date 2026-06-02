package service

import (
	"archive/zip"
	"bytes"
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestExtractText_PlainAndHTML(t *testing.T) {
	cases := []struct {
		name, file, in, want string
	}{
		{"txt", "a.txt", "hello  world", "hello world"},
		{"md", "a.md", "# Title\n\ntext", "# Title\n\ntext"},
		{"html", "a.html", "<p>hi <b>there</b></p>", "hi there"},
		{"unknown ext as utf8", "a.xyz", "raw bytes", "raw bytes"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := extractText(c.file, []byte(c.in))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Fatalf("got %q want %q", got, c.want)
			}
		})
	}
}

func TestExtractText_LegacyDocUnsupported(t *testing.T) {
	_, err := extractText("old.doc", []byte("anything"))
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("got %v want ErrUnsupportedFormat", err)
	}
}

func TestExtractDOCX(t *testing.T) {
	// Minimal OOXML body: two paragraphs, a tab and a run split across w:t.
	docXML := `<?xml version="1.0"?>
<w:document xmlns:w="x"><w:body>
<w:p><w:r><w:t>Hello</w:t></w:r><w:r><w:tab/></w:r><w:r><w:t>World</w:t></w:r></w:p>
<w:p><w:r><w:t>Second line</w:t></w:r></w:p>
</w:body></w:document>`

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(docXML)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := extractText("doc.docx", buf.Bytes())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "Hello World") {
		t.Fatalf("missing tab-joined run: %q", got)
	}
	if !strings.Contains(got, "Second line") {
		t.Fatalf("missing second paragraph: %q", got)
	}
}

func TestExtractDOCX_MissingDocumentXML(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	_, _ = zw.Create("other.xml")
	_ = zw.Close()

	_, err := extractText("bad.docx", buf.Bytes())
	if !errors.Is(err, ErrNoTextExtracted) {
		t.Fatalf("got %v want ErrNoTextExtracted", err)
	}
}

func TestExtractPDF_InvalidBytes(t *testing.T) {
	// Not a PDF → reader construction fails (any error is acceptable here).
	if _, err := extractText("garbage.pdf", []byte("not a pdf")); err == nil {
		t.Fatal("expected error for non-PDF bytes")
	}
}

func TestExtractPDF_HappyPath(t *testing.T) {
	got, err := extractText("hello.pdf", buildMinimalPDF("Hello PDF"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got, "Hello") {
		t.Fatalf("extracted text missing content: %q", got)
	}
}

// buildMinimalPDF assembles a tiny single-page, text-bearing PDF with a valid
// cross-reference table (offsets computed at write time, not hardcoded). Uses a
// standard-14 font (Helvetica) so no glyph-width tables are needed.
func buildMinimalPDF(text string) []byte {
	objs := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>",
		"<< /Length " + itoa(len("BT /F1 24 Tf 72 700 Td ("+text+") Tj ET")) + " >>\nstream\nBT /F1 24 Tf 72 700 Td (" + text + ") Tj ET\nendstream",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objs)+1)
	for i, body := range objs {
		offsets[i+1] = buf.Len()
		buf.WriteString(itoa(i+1) + " 0 obj\n" + body + "\nendobj\n")
	}

	xref := buf.Len()
	buf.WriteString("xref\n0 " + itoa(len(objs)+1) + "\n")
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= len(objs); i++ {
		buf.WriteString(pad10(offsets[i]) + " 00000 n \n")
	}
	buf.WriteString("trailer\n<< /Size " + itoa(len(objs)+1) + " /Root 1 0 R >>\n")
	buf.WriteString("startxref\n" + itoa(xref) + "\n%%EOF")
	return buf.Bytes()
}

func itoa(n int) string { return strconv.Itoa(n) }

func pad10(n int) string {
	s := strconv.Itoa(n)
	return strings.Repeat("0", 10-len(s)) + s
}
