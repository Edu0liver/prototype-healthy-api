package service

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"html"
	"io"
	"strings"

	"github.com/dslipak/pdf"
)

// extractText turns uploaded bytes into plain text based on the filename
// extension. Native: txt/md/html. PDF via dslipak/pdf, DOCX via OOXML zip parse
// (both pure-Go). Scanned/image PDFs yield no text → ErrNoTextExtracted so the
// limitation surfaces as document.status=failed. Legacy .doc is still
// unsupported (binary format, needs a converter).
func extractText(filename string, data []byte) (string, error) {
	ext := strings.ToLower(filepathExt(filename))
	switch ext {
	case ".txt", ".md", ".markdown", "":
		return normalize(string(data)), nil
	case ".html", ".htm":
		return normalize(stripHTML(string(data))), nil
	case ".pdf":
		return extractPDF(data)
	case ".docx":
		return extractDOCX(data)
	case ".doc":
		return "", ErrUnsupportedFormat
	default:
		// Best effort: treat as UTF-8 text.
		return normalize(string(data)), nil
	}
}

// extractPDF pulls plain text from a text-based PDF. Image-only (scanned) PDFs
// have no embedded text layer and return ErrNoTextExtracted (OCR out of scope).
func extractPDF(data []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	buf, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	raw, err := io.ReadAll(buf)
	if err != nil {
		return "", err
	}
	text := normalize(string(raw))
	if text == "" {
		return "", ErrNoTextExtracted
	}
	return text, nil
}

// extractDOCX reads an OOXML (.docx) file — a zip whose word/document.xml holds
// the body. Paragraphs (<w:p>) become newlines; runs (<w:t>) carry the text.
func extractDOCX(data []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	var doc *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			doc = f
			break
		}
	}
	if doc == nil {
		return "", ErrNoTextExtracted
	}
	rc, err := doc.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	var sb strings.Builder
	dec := xml.NewDecoder(rc)
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			// w:t carries run text; w:tab → space.
			if t.Name.Local == "tab" {
				sb.WriteByte(' ')
			}
			if t.Name.Local == "t" {
				var s string
				if err := dec.DecodeElement(&s, &t); err != nil {
					return "", err
				}
				sb.WriteString(s)
			}
		case xml.EndElement:
			// Paragraph / line break → newline.
			if t.Name.Local == "p" || t.Name.Local == "br" {
				sb.WriteByte('\n')
			}
		}
	}
	text := normalize(sb.String())
	if text == "" {
		return "", ErrNoTextExtracted
	}
	return text, nil
}

func stripHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, " ")
	return html.UnescapeString(s)
}

func normalize(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = whitespaceRe.ReplaceAllString(s, " ")
	s = blanklineRe.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

func filepathExt(name string) string {
	for i := len(name) - 1; i >= 0 && name[i] != '/'; i-- {
		if name[i] == '.' {
			return name[i:]
		}
	}
	return ""
}
