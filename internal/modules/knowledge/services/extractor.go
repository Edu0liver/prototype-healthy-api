package services

import (
	"html"
	"strings"
)

// extractText turns uploaded bytes into plain text based on the filename
// extension. v1 supports txt/md/html natively; PDF/DOCX need a parser library
// and currently return an error (status=failed) so the limitation is explicit.
func extractText(filename string, data []byte) (string, error) {
	ext := strings.ToLower(filepathExt(filename))
	switch ext {
	case ".txt", ".md", ".markdown", "":
		return normalize(string(data)), nil
	case ".html", ".htm":
		return normalize(stripHTML(string(data))), nil
	case ".pdf", ".docx", ".doc":
		return "", ErrUnsupportedFormat
	default:
		// Best effort: treat as UTF-8 text.
		return normalize(string(data)), nil
	}
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
