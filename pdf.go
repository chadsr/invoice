package main

import (
	"fmt"
	"image"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/signintech/gopdf"
)

const (
	quantityColumnOffset = 360
	rateColumnOffset     = 405
	amountColumnOffset   = 480
)

const (
	subtotalLabel = "Subtotal"
	discountLabel = "Discount"
	taxLabel      = "Tax"
	totalLabel    = "Total Due"
)

func writeAddress(pdf *gopdf.GoPdf, locX float64, rect *gopdf.Rect, fromAddress []string, cellOption *gopdf.CellOption) error {
	for _, v := range fromAddress {
		pdf.SetX(locX)
		err := pdf.CellWithOption(rect, v, *cellOption)
		if err != nil {
			return err
		}

		pdf.Br(18)
	}

	return nil
}

func writeDetails(pdf *gopdf.GoPdf, locX float64, rect *gopdf.Rect, details map[string]string, cellOption *gopdf.CellOption) error {
	keys := make([]string, 0, len(details))
	for k := range details {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		pdf.SetX(locX)
		err := pdf.CellWithOption(rect, fmt.Sprintf("%s: %s", k, details[k]), *cellOption)
		if err != nil {
			return err
		}

		pdf.Br(18)
	}

	return nil
}

func writeLogo(pdf *gopdf.GoPdf, logo string, logoSize float64) error {
	origX := pdf.GetX()
	origY := pdf.GetY()
	defer pdf.SetXY(origX, origY)

	if logo != "" {
		pdf.SetXY(gopdf.PageSizeA4.W-logoSize-pdf.MarginRight(), 0+pdf.MarginTop())
		width, height, err := getImageDimension(logo)
		if err != nil {
			return err
		}

		scaledWidth := logoSize
		scaledHeight := float64(height) * scaledWidth / float64(width)

		err = pdf.Image(logo, pdf.GetX(), pdf.GetY(), &gopdf.Rect{W: scaledWidth, H: scaledHeight})
		if err != nil {
			return err
		}
	}

	return nil
}

func writeTitle(pdf *gopdf.GoPdf, title, id, date string) error {
	err := pdf.SetFont("Inter-Bold", "", 24)
	if err != nil {
		return err
	}

	pdf.SetTextColor(0, 0, 0)
	err = pdf.Cell(nil, title)
	if err != nil {
		return err
	}

	pdf.Br(36)
	err = pdf.SetFont("Inter", "", 12)
	if err != nil {
		return err
	}

	pdf.SetTextColor(100, 100, 100)
	err = pdf.Cell(nil, "#")
	if err != nil {
		return err
	}

	err = pdf.Cell(nil, id)
	if err != nil {
		return err
	}

	pdf.SetTextColor(150, 150, 150)
	err = pdf.Cell(nil, "  ·  ")
	if err != nil {
		return err
	}

	pdf.SetTextColor(100, 100, 100)

	err = pdf.Cell(nil, date)
	if err != nil {
		return err
	}
	pdf.Br(12)
	return nil
}

func writeCompanyInfo(pdf *gopdf.GoPdf, locX float64, locY float64, name string, details map[string]string, address []string, cellOption gopdf.CellOption) error {
	pdf.SetXY(locX, locY)
	pdf.SetTextColor(75, 75, 75)

	formattedName := strings.ReplaceAll(name, `\n`, "\n")
	nameLines := strings.Split(formattedName, "\n")

	rectangle := gopdf.Rect{
		W: 100.0,
		H: 45.0,
	}

	for i := 0; i < len(nameLines); i++ {
		if i == 0 {
			err := pdf.SetFont("Inter-Bold", "", 12)
			if err != nil {
				return err
			}

			err = pdf.CellWithOption(&rectangle, nameLines[i], cellOption)
			if err != nil {
				return err
			}

			pdf.Br(18)
		} else {
			err := pdf.SetFont("Inter", "", 10)
			if err != nil {
				return err
			}

			err = pdf.CellWithOption(&rectangle, nameLines[i], cellOption)
			if err != nil {
				return err
			}

			pdf.Br(14)
		}
		pdf.SetX(locX)
	}

	err := pdf.SetFont("Inter", "", 10)
	if err != nil {
		return err
	}

	err = writeDetails(pdf, locX, &rectangle, details, &cellOption)
	if err != nil {
		return err
	}

	err = writeAddress(pdf, locX, &rectangle, address, &cellOption)
	if err != nil {
		return err
	}

	return nil
}

func writeTotalHours(pdf *gopdf.GoPdf, totalHours float64) error {
	err := pdf.SetFont("Inter", "", 9)
	if err != nil {
		return err
	}

	pdf.SetTextColor(75, 75, 75)
	pdf.SetX(rateColumnOffset)
	err = pdf.Cell(nil, "Total Hours")
	if err != nil {
		return err
	}

	pdf.SetTextColor(0, 0, 0)
	err = pdf.SetFontSize(11)
	if err != nil {
		return err
	}

	pdf.SetX(amountColumnOffset - 15)
	err = pdf.Cell(nil, fmt.Sprintf("%.2f", totalHours))
	if err != nil {
		return err
	}

	pdf.Br(24)
	return nil
}

func writeDueDate(pdf *gopdf.GoPdf, due string) error {
	err := pdf.SetFont("Inter", "", 9)
	if err != nil {
		return err
	}

	pdf.SetTextColor(75, 75, 75)
	pdf.SetX(rateColumnOffset)
	err = pdf.Cell(nil, "Due Date")
	if err != nil {
		return err
	}

	pdf.SetTextColor(0, 0, 0)
	err = pdf.SetFontSize(11)
	if err != nil {
		return err
	}

	pdf.SetX(amountColumnOffset - 15)
	err = pdf.Cell(nil, due)
	if err != nil {
		return err
	}

	pdf.Br(24)
	return nil
}

func writeNotes(pdf *gopdf.GoPdf, notes string) error {
	err := pdf.SetFont("Inter", "", 9)
	if err != nil {
		return err
	}

	pdf.SetTextColor(55, 55, 55)
	err = pdf.Cell(nil, "NOTES")
	if err != nil {
		return err
	}

	pdf.Br(18)
	err = pdf.SetFont("Inter", "", 9)
	if err != nil {
		return err
	}

	pdf.SetTextColor(0, 0, 0)

	formattedNotes := strings.ReplaceAll(notes, `\n`, "\n")
	notesLines := strings.Split(formattedNotes, "\n")

	for i := 0; i < len(notesLines); i++ {
		err = pdf.Cell(nil, notesLines[i])
		if err != nil {
			return err
		}
		pdf.Br(15)
	}

	pdf.Br(48)
	return nil
}

func writeFooter(pdf *gopdf.GoPdf, id string) error {
	pdf.SetY(800)

	err := pdf.SetFont("Inter", "", 10)
	if err != nil {
		return err
	}

	pdf.SetTextColor(55, 55, 55)
	err = pdf.Cell(nil, id)
	if err != nil {
		return err
	}

	pdf.SetStrokeColor(225, 225, 225)
	pdf.Line(pdf.GetX()+10, pdf.GetY()+6, 550, pdf.GetY()+6)
	pdf.Br(48)
	return nil
}

func writeItems(pdf *gopdf.GoPdf, width float64, currencySymbol string, items []string, quantities []float64, rates []float64, dates []time.Time) (totalCost float64, totalHours float64, err error) {
	itemsTable := pdf.NewTableLayout(pdf.GetX(), pdf.GetY(), 24, len(items))
	itemsTable.SetTableStyle(gopdf.CellStyle{
		BorderStyle: gopdf.BorderStyle{
			RGBColor: gopdf.RGBColor{R: 75, G: 75, B: 75},
			Top:      true,
			Bottom:   true,
		},
	})
	itemsTable.SetHeaderStyle(gopdf.CellStyle{
		BorderStyle: gopdf.BorderStyle{},
		Font:        "Inter-Bold",
		TextColor:   gopdf.RGBColor{R: 75, G: 75, B: 75},
		FontSize:    11,
	})

	itemsTable.SetCellStyle(gopdf.CellStyle{
		TextColor:   gopdf.RGBColor{R: 55, G: 55, B: 55},
		BorderStyle: gopdf.BorderStyle{},
		Font:        "Inter",
		FontSize:    8,
	})

	itemsTable.AddColumn("DATE", (width / 8), "center")
	itemsTable.AddColumn("DESCRIPTION", (width / 2), "center")
	itemsTable.AddColumn("HOURS", (width / 8), "center")
	itemsTable.AddColumn("RATE", (width / 8), "center")
	itemsTable.AddColumn("AMOUNT", (width / 8), "center")

	for i := range items {
		quantity := 1.0
		if len(quantities) > i {
			quantity = quantities[i]
		}

		rate := rates[0]
		if len(rates) > i {
			rate = rates[i]
		}

		total := float64(quantity) * rate
		itemsTable.AddRow([]string{
			dates[i].Format("02-01-2006"),
			items[i],
			strconv.FormatFloat(quantity, 'f', 2, 64),
			fmt.Sprintf("%s%.2f", currencySymbol, rate),
			fmt.Sprintf("%s%.2f", currencySymbol, total),
		})

		totalCost += float64(quantity) * rate
		totalHours += quantity
	}

	err = itemsTable.DrawTable()
	if err != nil {
		return
	}

	return
}

func writeTotal(pdf *gopdf.GoPdf, label string, total float64) error {
	err := pdf.SetFont("Inter", "", 9)
	if err != nil {
		return err
	}

	pdf.SetTextColor(75, 75, 75)
	pdf.SetX(rateColumnOffset)
	err = pdf.Cell(nil, label)
	if err != nil {
		return err
	}

	pdf.SetTextColor(0, 0, 0)
	err = pdf.SetFontSize(12)
	if err != nil {
		return err
	}

	pdf.SetX(amountColumnOffset - 15)
	if label == totalLabel {
		err = pdf.SetFont("Inter-Bold", "", 11.5)
		if err != nil {
			return err
		}
	}
	err = pdf.Cell(nil, currencySymbols[file.Currency]+strconv.FormatFloat(total, 'f', 2, 64))
	if err != nil {
		return err
	}

	pdf.Br(24)
	return nil
}

func writeTotals(pdf *gopdf.GoPdf, subtotal float64, discount float64, taxRate float64, taxInclusive bool, taxName string) error {
	if taxInclusive {
		err := writeTotal(pdf, subtotalLabel, subtotal-(subtotal*taxRate))
		if err != nil {
			return err
		}
	} else {
		err := writeTotal(pdf, subtotalLabel, subtotal)
		if err != nil {
			return err
		}
	}

	if taxRate > 0 {
		err := writeTotal(pdf, fmt.Sprintf("%s %.0f%%", taxName, taxRate*100), subtotal*taxRate)
		if err != nil {
			return err
		}
	}
	if discount > 0 {
		err := writeTotal(pdf, discountLabel, discount)
		if err != nil {
			return err
		}
	}

	total := subtotal - discount
	if !taxInclusive && taxRate > 0 {
		total += (subtotal * taxRate)
	}

	err := writeTotal(pdf, totalLabel, total)
	if err != nil {
		return err
	}

	return nil
}

func getImageDimension(imagePath string) (int, int, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}
	return image.Width, image.Height, nil
}
