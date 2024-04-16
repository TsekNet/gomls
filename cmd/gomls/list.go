package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"gomls/pkg/helpers"
	"gomls/pkg/zillow"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/google/subcommands"
)

const (
	purple    = lipgloss.Color("99")
	gray      = lipgloss.Color("245")
	lightGray = lipgloss.Color("241")
	layout    = "2006-01-02"
)

var (
	errLocation = errors.New("required flag -location is missing")
	re          = lipgloss.NewRenderer(os.Stdout)

	// HeaderStyle is the lipgloss style used for the table headers.
	HeaderStyle = re.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center).Background(gray)
	// CellStyle is the base lipgloss style used for the table rows.
	CellStyle = re.NewStyle().Padding(0, 1).Width(14)
	// OddRowStyle is the lipgloss style used for odd-numbered table rows.
	OddRowStyle = CellStyle.Copy().Foreground(gray)
	// EvenRowStyle is the lipgloss style used for even-numbered table rows.
	EvenRowStyle = CellStyle.Copy().Foreground(lightGray)
	// BorderStyle is the lipgloss style used for the table border.
	BorderStyle = lipgloss.NewStyle().Foreground(purple)
)

type ListCmd struct {
	beds, baths, price              int
	output, location, property_type string
	sold                            bool
}

type houseData struct {
	Houses       helpers.HouseSlice
	At, Location string
	Sold         bool
}

func (ListCmd) Name() string     { return "list" }
func (ListCmd) Synopsis() string { return "Lists items with optional output format" }
func (ListCmd) Usage() string {
	return fmt.Sprintf("%s list\n", filepath.Base(os.Args[0]))
}

func (l *ListCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&l.output, "output", "plain", "Output format\nMust be one of: [plain, table, json, html, csv]")
	f.StringVar(&l.location, "location", "10001", "Filter by location of the properties (can be neighborhood, zip code, etc.).\nType this into your search bar on zillow.com if you want to confirm the format.")
	f.StringVar(&l.property_type, "property_type", "", "Filter by property type\nMust be one of: [APARTMENT, CONDO, MULTI_FAMILY, SINGLE_FAMILY]")
	f.BoolVar(&l.sold, "sold", false, "Filter by sold properties")
	f.IntVar(&l.baths, "baths", 0, "Filter by number of baths")
	f.IntVar(&l.beds, "beds", 0, "Filter by number of beds")
	f.IntVar(&l.price, "price", 0, "Filter by price")
}

func (l ListCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...any) subcommands.ExitStatus {
	if l.location == "" {
		fmt.Println(errLocation)
		return subcommands.ExitUsageError
	}

	d := helpers.Details{
		Baths:        l.baths,
		Beds:         l.beds,
		Price:        l.price,
		Location:     l.location,
		PropertyType: l.property_type,
		Sold:         l.sold,
	}

	switch l.output {
	case "html":
		err := outputHTML(d)
		if err != nil {
			fmt.Fprintf(f.Output(), "Error generating HTML output: %v\n", err)
			return subcommands.ExitFailure
		}
	case "csv":
		err := outputCSV(d)
		if err != nil {
			fmt.Fprintf(f.Output(), "Error generating CSV output: %v\n", err)
			return subcommands.ExitFailure
		}
	case "plain":
		outputPlain(d)
	case "table":
		outputTable(d)
	case "json":
		err := outputJSON(d)
		if err != nil {
			fmt.Fprintf(f.Output(), "Error generating CSV output: %v\n", err)
			return subcommands.ExitFailure
		}
	default:
		// Invalid format handled as before
		fmt.Fprintf(f.Output(), "Invalid output format: %s\n", l.output)
		return subcommands.ExitUsageError
	}

	return subcommands.ExitSuccess
}

func outputHTML(d helpers.Details) error {
	houses := zillow.Query(d)

	if len(houses) == 0 {
		return nil
	}

	file := filepath.Join(os.Getenv("TEMP"), "listings.html")

	data := houseData{
		Houses:   houses,
		At:       time.Now().Format(layout),
		Location: d.Location,
		Sold:     d.Sold,
	}

	// Marshal housesWithLink to JSON
	_, err := json.Marshal(houses)
	if err != nil {
		return fmt.Errorf("failed to marshal houses to JSON: %w", err)
	}

	// Parse the template
	t := template.New("listings.html")
	t.Parse(helpers.TemplateHTML)

	// Create and write HTML file
	htmlFile, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer htmlFile.Close()

	err = t.Execute(htmlFile, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	fmt.Printf("Opening %q in your browser...\n", file)
	helpers.OpenBrowser(file)

	return nil
}

func outputPlain(d helpers.Details) {
	houses := zillow.Query(d)

	header := lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("254"))
	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("#555"))
	fieldfmt := lipgloss.NewStyle()

	fmt.Println(header.Render("House Listings:\n"))
	for _, house := range houses {
		typeOfHouse := reflect.TypeOf(house)
		for i := 0; i < typeOfHouse.NumField(); i++ {
			field := typeOfHouse.Field(i)
			fieldName := field.Name
			fieldValue := reflect.ValueOf(house).Field(i).Interface()

			if fieldValue != "" {
				if fieldName == "Address" {
					fmt.Println(header.Render(fmt.Sprintf("%v", fieldValue)))
				} else {
					fmt.Println(separator.Render("- "+fieldName+": ") + fieldfmt.Render(fmt.Sprintf("%v", fieldValue)))
				}
			}
		}

		fmt.Println()
	}
}

func outputTable(d helpers.Details) {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderRow(true).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return HeaderStyle
			case row%2 == 0:
				return EvenRowStyle
			default:
				return OddRowStyle
			}
		}).
		Headers(helpers.StructToSlice(helpers.House{})...).
		Rows(helpers.SliceToRow(zillow.Query(d))...)

	if t == nil {
		return
	}

	fmt.Println(t)
}

func outputJSON(d helpers.Details) error {
	s, _ := json.MarshalIndent(zillow.Query(d), "", "\t")

	if s == nil {
		return nil
	}

	file := filepath.Join(os.Getenv("TEMP"), "listings.json")
	// output to a JSON file called "filepath.Join(os.Getenv("TEMP"), "listings.html")"
	if err := os.WriteFile(file, s, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	fmt.Printf("Wrote JSON file: %q\n", file)

	return nil
}

func outputCSV(d helpers.Details) error {
	houses := zillow.Query(d)

	if len(houses) == 0 {
		return nil
	}

	typeOfHouse := reflect.TypeOf(houses[0])
	header := make([]string, typeOfHouse.NumField())
	for i := 0; i < typeOfHouse.NumField(); i++ {
		field := typeOfHouse.Field(i)
		header[i] = field.Name
	}

	file, err := os.Create(filepath.Join(os.Getenv("TEMP"), "listings.csv"))
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write(header)

	for _, house := range houses {
		typeOfHouse := reflect.TypeOf(house)
		row := make([]string, typeOfHouse.NumField())
		for i := 0; i < typeOfHouse.NumField(); i++ {
			fieldValue := reflect.ValueOf(house).Field(i).Interface()
			row[i] = fmt.Sprintf("%v", fieldValue)
		}

		writer.Write(row)
	}

	fmt.Printf("Wrote CSV file: %q\n", file.Name())

	return nil
}
