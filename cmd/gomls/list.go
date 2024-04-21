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
	"github.com/google/subcommands"
)

const (
	layout = "2006-01-02"
)

var (
	errLocation = errors.New("required flag -location is missing")
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
	f.StringVar(&l.output, "output", "plain", "Output format\nMust be one of: [plain, json, html, csv]")
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
	case "json":
		err := outputJSON(d)
		if err != nil {
			fmt.Fprintf(f.Output(), "Error generating JSON output: %v\n", err)
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

	fmt.Println(header.Render("\nHouse Listings:"))

	for _, house := range houses {
		fmt.Println(header.Render(fmt.Sprintf("\n%v", house.Property.Address)))

		// Top-level properties
		// Prices
		if house.Property.ListPrice > 0 {
			fmt.Printf("%s $%d\n", separator.Render("- List Price:"), house.Property.ListPrice)
		}
		if house.Property.SoldPrice > 0 {
			fmt.Printf("%s $%d\n", separator.Render("- Sold Price:"), house.Property.SoldPrice)
		}
		if house.Property.PriceDiff > 0 {
			fmt.Printf("%s $%d\n", separator.Render("- Price Diff:"), house.Property.PriceDiff)
		}
		if house.Property.PriceDiffPercent > 0 {
			fmt.Printf("%s %d%%\n", separator.Render("- Price Diff %:"), int64(house.Property.PriceDiffPercent))
		}
		if house.Property.ListDate != "" {
			fmt.Printf("%s %s\n", separator.Render("- List Date:"), house.Property.ListDate)
		}
		if house.Property.SoldDate != "" {
			fmt.Printf("%s %s\n", separator.Render("- Sold Date:"), house.Property.SoldDate)
		}

		// Misc
		if house.Property.DatePostedString != "" {
			fmt.Printf("%s %s\n", separator.Render("- Date Posted:"), house.Property.DatePostedString)
		}
		if house.Property.HomeStatus != "" {
			fmt.Printf("%s %s\n", separator.Render("- Home Status:"), house.Property.HomeStatus)
		}
		if house.Property.HomeType != "" {
			fmt.Printf("%s %s\n", separator.Render("- Home Type:"), house.Property.HomeType)
		}
		if house.Property.Description != "" {
			fmt.Printf("%s %s\n", separator.Render("- Description:"), house.Property.Description)
		}

		// Links
		fmt.Println(separator.Render("- Links:"))
		if house.Property.FullUrl != "" {
			fmt.Printf("  %s %s\n", separator.Render("- Full URL:"), house.Property.FullUrl)
		}
		if house.Property.MapsUrl != "" {
			fmt.Printf("  %s %s\n", separator.Render("- Maps URL:"), house.Property.MapsUrl)
		}
		if house.Property.DesktopWebHdpImageLink != "" {
			fmt.Printf("  %s %s\n", separator.Render("- Image Link:"), house.Property.DesktopWebHdpImageLink)
		}

		// Facts
		fmt.Println(separator.Render("- Facts:"))
		if house.Property.ResoFacts.Bedrooms > 0 {
			fmt.Printf("  %s %d\n", separator.Render("- Bedrooms:"), house.Property.ResoFacts.Bedrooms)
		}
		if house.Property.ResoFacts.Bathrooms > 0 {
			fmt.Printf("  %s %d\n", separator.Render("- Bathrooms:"), house.Property.ResoFacts.Bathrooms)
		}
		if house.Property.ResoFacts.LivingArea != "" {
			fmt.Printf("  %s %s\n", separator.Render("- Living Area:"), house.Property.ResoFacts.LivingArea)
		}

		// Open house
		if len(house.Property.OpenHouseSchedule) > 0 {
			fmt.Println(separator.Render("- Showing:"))
			for _, oh := range house.Property.OpenHouseSchedule {
				fmt.Printf("  %s %v -> %v\n", separator.Render("-"), oh.StartTime, oh.EndTime)
			}
		}

		// Price history
		if len(house.Property.PriceHistory) > 0 {
			fmt.Println(separator.Render("- Price History:"))
			for _, ph := range house.Property.PriceHistory {
				fmt.Printf("  %s Date: %v, Event: %v, Price: $%v\n", separator.Render("-"), ph.Date, ph.Event, ph.Price)
			}
		}

	}
}

func outputJSON(d helpers.Details) error {
	s, _ := json.MarshalIndent(zillow.Query(d), "", "\t")

	if s == nil {
		return nil
	}

	file := filepath.Join(os.Getenv("TEMP"), "listings.json")
	// output to a JSON file
	if err := os.WriteFile(file, s, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	fmt.Printf("Wrote JSON file: %q\n", file)

	return nil
}

// TODO: Fix CSV formats
func outputCSV(d helpers.Details) error {
	houses := zillow.Query(d)

	if len(houses) == 0 {
		return nil
	}

	typeOfHouse := reflect.TypeOf(houses[0].Property)
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
		typeOfHouse := reflect.TypeOf(house.Property)
		row := make([]string, typeOfHouse.NumField())
		for i := 0; i < typeOfHouse.NumField(); i++ {
			fieldValue := reflect.ValueOf(house.Property).Field(i).Interface()
			row[i] = fmt.Sprintf("%v", fieldValue)
		}

		writer.Write(row)
	}

	fmt.Printf("Wrote CSV file: %q\n", file.Name())

	return nil
}
