package main

import (
	_ "embed"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/signintech/gopdf"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed "Inter/Inter Variable/Inter.ttf"
var interFont []byte

//go:embed "Inter/Inter Hinted for Windows/Desktop/Inter-Bold.ttf"
var interBoldFont []byte

type Invoice struct {
	Id       string `json:"id" yaml:"id"`
	IdPrefix string `json:"idPrefix" yaml:"idPrefix"`

	Website string `json:"website" yaml:"website"`
	Email   string `json:"email" yaml:"email"`

	Title string `json:"title" yaml:"title"`

	Logo     string  `json:"logo" yaml:"logo"`
	LogoSize float64 `json:"logoSize" yaml:"logoSize"`

	From        string            `json:"from" yaml:"from"`
	FromDetails map[string]string `json:"fromDetails" yaml:"fromDetails"`
	FromAddress []string          `json:"fromAddress" yaml:"fromAddress"`

	To        string            `json:"to" yaml:"to"`
	ToDetails map[string]string `json:"toDetails" yaml:"toDetails"`
	ToAddress []string          `json:"toAddress" yaml:"toAddress"`

	Dates   []time.Time `json:"dates" yaml:"dates"`
	Date    string      `json:"date" yaml:"date"`
	Due     string      `json:"due" yaml:"due"`
	DueDays int         `json:"dueDays" yaml:"dueDays"`

	Items             []string  `json:"items" yaml:"items"`
	Quantities        []float64 `json:"quantities" yaml:"quantities"`
	Rates             []float64 `json:"rates" yaml:"rates"`
	RatesTaxInclusive bool      `json:"ratesTaxInclusive" yaml:"ratesTaxInclusive"`

	Tax      float64 `json:"tax" yaml:"tax"`
	TaxName  string  `json:"taxName" yaml:"taxName"`
	Discount float64 `json:"discount" yaml:"discount"`
	Currency string  `json:"currency" yaml:"currency"`

	Note string `json:"note" yaml:"note"`
}

type WorklogEntry struct {
	Date        time.Time
	Hours       float64
	Description string
}

func RoundTo(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}

func readWorklogCsv(filePath string) ([]WorklogEntry, error) {
	entries := make([]WorklogEntry, 0)

	f, err := os.Open(filePath)
	if err != nil {
		return entries, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return entries, err
	}

	expHeaders := []string{"Date", "Start", "End", "Worked", "Titles"}
	for i, record := range records {
		if i == 0 {
			for i, header := range record {
				if header != expHeaders[i] {
					return entries, fmt.Errorf("unexpected worklog headers")
				}
			}
			continue
		}

		entryDate, err := time.Parse("2006-01-02", record[0])
		if err != nil {
			return entries, err
		}

		durationValues := strings.Split(record[3], ":")
		durationHours, err := strconv.Atoi(durationValues[0])
		if err != nil {
			return entries, err
		}

		durationMinutes, err := strconv.Atoi(durationValues[1])
		if err != nil {
			return entries, err
		}

		durationTotal := float64(durationHours) + (float64(durationMinutes) / 60.0)

		entryDescription := record[4]

		entry := WorklogEntry{
			Date:        entryDate,
			Hours:       durationTotal,
			Description: entryDescription,
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func DefaultInvoice() Invoice {
	return Invoice{
		Id:                time.Now().Format("200601"),
		IdPrefix:          "INV",
		Email:             "some@business.com",
		Website:           "business.com",
		Title:             "INVOICE",
		LogoSize:          100,
		Rates:             []float64{25},
		Quantities:        []float64{2},
		Dates:             []time.Time{time.Now()},
		Items:             []string{"Paper Cranes"},
		From:              "Project Folded, Inc.",
		FromDetails:       map[string]string{},
		FromAddress:       []string{"1", "Main st", "Newyark", "626112"},
		To:                "Untitled Corporation, Inc.",
		ToDetails:         map[string]string{},
		ToAddress:         []string{"1/56A", "Main st", "Newyark", "626112"},
		Date:              time.Now().Format("Jan 02, 2006"),
		Due:               time.Now().AddDate(0, 0, 14).Format("Jan 02, 2006"),
		DueDays:           14,
		RatesTaxInclusive: false,
		Tax:               0,
		TaxName:           "VAT",
		Discount:          0,
		Currency:          "USD",
	}
}

var (
	configPath     string
	worklogPath    string
	output         string
	file           = Invoice{}
	defaultInvoice = DefaultInvoice()
)

func init() {
	viper.AutomaticEnv()

	generateCmd.Flags().StringVar(&configPath, "config", "", "Config file (.json/.yaml)")
	generateCmd.Flags().StringVar(&worklogPath, "worklog", "", "Worklog file (.csv)")

	generateCmd.Flags().StringVar(&file.Id, "id", defaultInvoice.Id, "ID")
	generateCmd.Flags().StringVar(&file.IdPrefix, "id-prefix", defaultInvoice.IdPrefix, "ID Prefix")
	generateCmd.Flags().StringVar(&file.Title, "title", "INVOICE", "Title")
	generateCmd.Flags().StringVar(&file.Website, "website", defaultInvoice.Website, "Website")
	generateCmd.Flags().StringVar(&file.Email, "email", defaultInvoice.Email, "Email address")

	generateCmd.Flags().Float64SliceVarP(&file.Rates, "rate", "r", defaultInvoice.Rates, "Rates")
	generateCmd.Flags().Float64SliceVarP(&file.Quantities, "quantity", "q", defaultInvoice.Quantities, "Quantities")
	generateCmd.Flags().StringSliceVarP(&file.Items, "item", "i", defaultInvoice.Items, "Items")

	generateCmd.Flags().StringVarP(&file.Logo, "logo", "l", defaultInvoice.Logo, "Company logo")
	generateCmd.Flags().Float64Var(&file.LogoSize, "logo-size", defaultInvoice.LogoSize, "Logo size")

	generateCmd.Flags().StringVarP(&file.From, "from", "f", defaultInvoice.From, "Issuing company")
	generateCmd.Flags().StringSliceVar(&file.FromAddress, "from-address", defaultInvoice.FromAddress, "Issuing company address")

	generateCmd.Flags().StringVarP(&file.To, "to", "t", defaultInvoice.To, "Recipient company")
	generateCmd.Flags().StringSliceVarP(&file.ToAddress, "to-address", "a", defaultInvoice.ToAddress, "Receiving company address")

	generateCmd.Flags().StringVar(&file.Date, "date", defaultInvoice.Date, "Date")
	generateCmd.Flags().StringVar(&file.Due, "due", defaultInvoice.Due, "Payment due date")
	generateCmd.Flags().IntVar(&file.DueDays, "due-days", defaultInvoice.DueDays, "Payment due days after generation date")

	generateCmd.Flags().BoolVar(&file.RatesTaxInclusive, "rates-tax-inclusive", defaultInvoice.RatesTaxInclusive, "Rates tax inclusive")
	generateCmd.Flags().Float64Var(&file.Tax, "tax", defaultInvoice.Tax, "Tax")
	generateCmd.Flags().StringVar(&file.TaxName, "tax-name", defaultInvoice.TaxName, "Tax Name")
	generateCmd.Flags().Float64VarP(&file.Discount, "discount", "d", defaultInvoice.Discount, "Discount")
	generateCmd.Flags().StringVarP(&file.Currency, "currency", "c", defaultInvoice.Currency, "Currency")

	generateCmd.Flags().StringVarP(&file.Note, "note", "n", "", "Note")
	generateCmd.Flags().StringVarP(&output, "output", "o", "invoice.pdf", "Output file (.pdf)")

	flag.Parse()
}

var rootCmd = &cobra.Command{
	Use:   "invoice",
	Short: "Invoice generates invoices from the command line.",
	Long:  `Invoice generates invoices from the command line.`,
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an invoice",
	Long:  `Generate an invoice`,
	RunE: func(cmd *cobra.Command, args []string) error {

		if configPath != "" {
			err := importData(configPath, &file, cmd.Flags())
			if err != nil {
				return err
			}
		}

		worklogEntries, err := readWorklogCsv(worklogPath)
		if err != nil {
			return err
		}

		if len(worklogEntries) > 0 {
			file.Items = make([]string, len(worklogEntries))
			file.Quantities = make([]float64, len(worklogEntries))
			file.Dates = make([]time.Time, len(worklogEntries))

			for i, entry := range worklogEntries {
				file.Items[i] = entry.Description
				file.Quantities[i] = entry.Hours
				file.Dates[i] = entry.Date
			}
		}

		file.Id = fmt.Sprintf("%s%s", file.IdPrefix, file.Id)

		pdf := gopdf.GoPdf{}
		pdf.Start(gopdf.Config{
			PageSize: *gopdf.PageSizeA4,
			Protection: gopdf.PDFProtectionConfig{
				UseProtection: true,
				Permissions:   gopdf.PermissionsPrint,
			},
		})
		pdf.SetMargins(40, 40, 40, 40)
		pdf.AddPage()

		pageWidth := gopdf.PageSizeA4.W - pdf.MarginLeft() - pdf.MarginRight()
		// pageHeight := gopdf.PageSizeA4.H - pdf.MarginTop() - pdf.MarginBottom()

		err = pdf.AddTTFFontData("Inter", interFont)
		if err != nil {
			return err
		}

		err = pdf.AddTTFFontData("Inter-Bold", interBoldFont)
		if err != nil {
			return err
		}

		err = writeTitle(&pdf, file.Title, file.Id, file.Date)
		if err != nil {
			return err
		}

		err = writeLogo(&pdf, file.Logo, file.LogoSize)
		if err != nil {
			return err
		}

		pdf.Br(24)

		currentX := pdf.GetX()
		currentY := pdf.GetY()

		err = writeCompanyInfo(&pdf, currentX, currentY, file.To, file.ToDetails, file.ToAddress, gopdf.CellOption{})
		if err != nil {
			return err
		}

		err = writeCompanyInfo(&pdf, pageWidth-60, currentY, file.From, file.FromDetails, file.FromAddress, gopdf.CellOption{Align: gopdf.Right})
		if err != nil {
			return err
		}

		pdf.Br(18)

		totalCost, totalHours, err := writeItems(&pdf, pageWidth, currencySymbols[file.Currency], file.Items, file.Quantities, file.Rates, file.Dates)
		if err != nil {
			return err
		}

		pdf.Br(32)

		xPos := pdf.GetX()
		yPos := pdf.GetY()

		if file.Note != "" {
			err = writeNotes(&pdf, file.Note)
			if err != nil {
				return err
			}
		}

		pdf.SetXY(xPos, yPos)

		err = writeTotalHours(&pdf, totalHours)
		if err != nil {
			return err
		}

		err = writeTotals(&pdf, totalCost, totalCost*file.Discount, file.Tax, file.RatesTaxInclusive, file.TaxName)
		if err != nil {
			return err
		}

		if file.Due != "" {
			err = writeDueDate(&pdf, file.Due)
			if err != nil {
				return err
			}
		} else if file.DueDays > 0 {
			dueDate := time.Now().AddDate(0, 0, file.DueDays).Format("Jan 02, 2006")
			err = writeDueDate(&pdf, dueDate)
			if err != nil {
				return err
			}
		}

		err = writeFooter(&pdf, file.Id)
		if err != nil {
			return err
		}

		output = fmt.Sprintf("%s_%s.pdf", strings.TrimSuffix(output, ".pdf"), file.Id)
		err = pdf.WritePdf(output)
		if err != nil {
			return err
		}

		fmt.Printf("Generated %s\n", output)

		return nil
	},
}

func main() {
	rootCmd.AddCommand(generateCmd)
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
