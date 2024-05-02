package main

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tealeg/xlsx"
)

func main() {
	baseURL := "https://kolesa.kz/cars/chevrolet/camaro/"
	page := 1

	// Create a new Excel file
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Prices")
	if err != nil {
		fmt.Println("Error creating Excel sheet:", err)
		return
	}

	var allEntries []entry // Slice to hold all entries (price, year)

	for {
		url := baseURL
		if page > 1 {
			url = fmt.Sprintf("%s?page=%d", baseURL, page)
		}

		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error making HTTP request:", err)
			return
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			fmt.Println("Error parsing HTML:", err)
			return
		}

		entriesAdded := false // Flag to track if any entries were added from this page

		doc.Find(".a-card").Each(func(i int, s *goquery.Selection) {
			price := strings.TrimSpace(s.Find(".a-card__price").Text())
			price = strings.ReplaceAll(price, "&nbsp;", "")
			price = strings.ReplaceAll(price, "₸", "")

			// Preprocess price: remove non-numeric characters and convert to integer
			price = strings.ReplaceAll(price, " ", "")      // Remove spaces
			price = strings.ReplaceAll(price, "\u00a0", "") // Remove non-breaking spaces
			priceInt, err := strconv.Atoi(price)
			if err != nil {
				fmt.Println("Error converting price to integer:", err)
				return
			}

			description := strings.TrimSpace(s.Find(".a-card__description").Text())
			year := description[:4] // Extract first 4 characters as the year

			allEntries = append(allEntries, entry{price: priceInt, year: year})
			entriesAdded = true
		})

		// Check if any entries were added from this page
		if !entriesAdded {
			break
		}

		page++
	}

	// Sort entries by price
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].price < allEntries[j].price
	})

	// Write sorted entries to Excel
	for _, e := range allEntries {
		row := sheet.AddRow()
		cell := row.AddCell()
		cell.SetInt(e.price) // Set integer value
		cell = row.AddCell()
		cell.Value = e.year
	}

	// Save the Excel file
	err = file.Save("prices_sorted.xlsx")
	if err != nil {
		fmt.Println("Error saving Excel file:", err)
		return
	}

	fmt.Println("Excel file saved successfully.")
}

type entry struct {
	price int // Changed type to int for sorting
	year  string
}
