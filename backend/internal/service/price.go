package service

import (
	"fmt"
	"math"
)

var PaperSizeMultiplier = map[string]float64{
	"a4": 1,
	"f4": 1.1,
	"a3": 2,
}

const EstimatedMinutesBase = 30

type Pricing struct {
	PerPage        int64 `json:"perPage"`
	Subtotal       int64 `json:"subtotal"`
	FinishingPrice int64 `json:"finishingPrice"`
	Total          int64 `json:"total"`
	Formatted      string `json:"formatted"`
}

func CalculatePrintPrice(pricePerPage, pageCount, copies, finishingPrice int64, paperSize string) Pricing {
	mul := PaperSizeMultiplier[paperSize]
	if mul == 0 {
		mul = 1
	}
	perPage := int64(math.Round(float64(pricePerPage) * mul))
	subtotal := perPage * pageCount * copies
	total := subtotal + finishingPrice
	return Pricing{PerPage: perPage, Subtotal: subtotal, FinishingPrice: finishingPrice, Total: total, Formatted: FormatRupiah(total)}
}

func EstimateMinutes(pageCount, copies int64) int64 {
	return EstimatedMinutesBase + ((pageCount*copies+9)/10)*5
}

func FormatRupiah(v int64) string {
	s := fmt.Sprintf("%d", v)
	n := len(s)
	withSep := ""
	for i, ch := range s {
		withSep += string(ch)
		rem := n - 1 - i
		if rem > 0 && rem%3 == 0 {
			withSep += "."
		}
	}
	return "Rp " + withSep
}