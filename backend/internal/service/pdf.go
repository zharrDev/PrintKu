package service

import (
	"archive/zip"
	"bytes"
	"io"
	"regexp"
	"strings"
)

const maxPages = 1000
const wordsPerPage = 350

var (
	rePage  = regexp.MustCompile(`(?i)/Type\s*/Page[^s]`)
	rePages = regexp.MustCompile(`(?i)/Type\s*/Pages`)
	reCount = regexp.MustCompile(`(?i)/Count\s+(\d+)`)
	reTag   = regexp.MustCompile(`<[^>]+>`)
)

// CountPdfPages menghitung jumlah halaman PDF berbasis struktur objek.
// Heuristik: jumlah "/Type /Page" (bukan /Pages) dikurangi,
// fallback ke nilai /Count pada objek /Pages, lalu fallback ke 1.
func CountPdfPages(data []byte) int {
	pages := len(rePage.FindAll(data, -1))
	pagesCollections := len(rePages.FindAll(data, -1))
	n := pages - pagesCollections
	if n > 0 {
		if n > maxPages {
			return maxPages
		}
		return n
	}
	if m := reCount.FindSubmatch(data); m != nil {
		v := parseNum(m[1])
		if v > 0 {
			return min(v, maxPages)
		}
	}
	if pages > 0 {
		return min(pages, maxPages)
	}
	return 1
}

// CountDocxPages mengestimasi jumlah halaman DOCX dengan cara yang lebih akurat:
// 1. Buka file DOCX sebagai ZIP
// 2. Baca word/document.xml
// 3. Hitung kata dari isi XML
// 4. Estimasi halaman berdasarkan jumlah kata
func CountDocxPages(data []byte) int {
	text := extractDocxText(data)
	if text == "" {
		return 1
	}
	words := len(strings.Fields(text))
	pages := (words + wordsPerPage - 1) / wordsPerPage
	if pages < 1 {
		pages = 1
	}
	if pages > maxPages {
		return maxPages
	}
	return pages
}

// extractDocxText membaca isi text dari file DOCX (ZIP format).
// DOCX adalah file ZIP yang berisi word/document.xml.
func extractDocxText(data []byte) string {
	// Parse sebagai ZIP
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		// Fallback: strip tags dari raw data
		return reTag.ReplaceAllString(string(data), " ")
	}

	// Cari word/document.xml dalam ZIP
	for _, f := range zipReader.File {
		if f.Name == "word/document.xml" {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			defer rc.Close()
			content, err := io.ReadAll(rc)
			if err != nil {
				continue
			}
			// Strip XML tags untuk mendapatkan text murni
			text := reTag.ReplaceAllString(string(content), " ")
			return text
		}
	}

	// Fallback: strip tags dari raw data
	return reTag.ReplaceAllString(string(data), " ")
}

// CountPages untuk ekstensi tertentu.
// CATATAN: .doc (bukan .docx) tidak didukung karena format binary lama.
// Jika diperlukan, gunakan LibreOffice headless untuk konversi ke .docx.
func CountPages(ext string, data []byte) int {
	switch strings.ToLower(ext) {
	case "pdf":
		return CountPdfPages(data)
	case "docx":
		return CountDocxPages(data)
	case "jpg", "jpeg", "png":
		return 1
	default:
		return 1
	}
}

func parseNum(b []byte) int {
	n := 0
	b = bytes.TrimSpace(b)
	for _, c := range b {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
