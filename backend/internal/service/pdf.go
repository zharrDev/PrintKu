package service

import (
	"bytes"
	"regexp"
	"strings"
)

const maxPages = 1000
const wordsPerPage = 350

var (
	rePage  = regexp.MustCompile(`(?i)/Type\s*/Page`)
	rePages = regexp.MustCompile(`(?i)/Type\s*/Pages`)
	reCount = regexp.MustCompile(`(?i)/Count\s+(\d+)`)
	reTag   = regexp.MustCompile(`<[^>]+>`)
)

// CountPdfPages menghitung jumlah halaman PDF berbasis struktur objek.
// Heuristik: jumlah "/Type /Page" dikurangi "/Type /Pages", fallback ke
// nilai /Count pada objek /Pages, lalu jumlah objek halaman.
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

// CountDocxPages mengestimasi jumlah halaman DOCX dari jumlah kata.
func CountDocxPages(data []byte) int {
	text := reTag.ReplaceAllString(string(data), " ")
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

// CountPages untuk ekstensi tertentu (pdf/docx/jpg/jpeg/png/dll).
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