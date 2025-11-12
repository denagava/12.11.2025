package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

type PDFGenerator struct{}

func NewPDFGenerator() *PDFGenerator {
	return &PDFGenerator{}
}

func (pg *PDFGenerator) GeneratePDF(linkSets map[int]*LinkSet) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Link Status Report")
	pdf.Ln(12)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(20, 10, "Set ID")
	pdf.Cell(80, 10, "URL")
	pdf.Cell(30, 10, "Status")
	pdf.Cell(40, 10, "Timestamp")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 10)
	for linksNum, linkSet := range linkSets {
		for url, status := range linkSet.Links {
			displayURL := url
			if len(displayURL) > 50 {
				displayURL = displayURL[:47] + "..."
			}
			pdf.Cell(20, 8, fmt.Sprintf("%d", linksNum))
			pdf.Cell(80, 8, displayURL)
			if strings.ToLower(status) == "available" {
				pdf.SetTextColor(0, 128, 0)
			} else {
				pdf.SetTextColor(255, 0, 0)
			}
			pdf.Cell(30, 8, status)
			pdf.SetTextColor(0, 0, 0)

			pdf.Cell(40, 8, linkSet.Timestamp.Format("2006-01-02 15:04:05"))
			pdf.Ln(8)
		}
		pdf.Ln(2)
	}
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
