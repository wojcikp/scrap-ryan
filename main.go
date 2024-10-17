package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type FlightResponse struct {
	ArrivalAirportCategories interface{} `json:"arrivalAirportCategories"`
	Fares                    []Fare      `json:"fares"`
	NextPage                 *string     `json:"nextPage"` // Assuming nextPage can be null
	Size                     int         `json:"size"`
}

type Fare struct {
	Outbound Outbound `json:"outbound"`
	Summary  Summary  `json:"summary"`
}

type Outbound struct {
	DepartureAirport Airport  `json:"departureAirport"`
	ArrivalAirport   Airport  `json:"arrivalAirport"`
	DepartureDate    string   `json:"departureDate"`
	ArrivalDate      string   `json:"arrivalDate"`
	Price            Price    `json:"price"`
	FlightKey        string   `json:"flightKey"`
	FlightNumber     string   `json:"flightNumber"`
	PreviousPrice    *float64 `json:"previousPrice"` // Assuming previousPrice can be null
	PriceUpdated     int64    `json:"priceUpdated"`
}

type Airport struct {
	CountryName string `json:"countryName"`
	IATACode    string `json:"iataCode"`
	Name        string `json:"name"`
	SEOName     string `json:"seoName"`
	City        City   `json:"city"`
}

type City struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	MacCode     string `json:"macCode,omitempty"` // Optional
	CountryCode string `json:"countryCode"`
}

type Price struct {
	Value               float64 `json:"value"`
	ValueMainUnit       string  `json:"valueMainUnit"`
	ValueFractionalUnit string  `json:"valueFractionalUnit"`
	CurrencyCode        string  `json:"currencyCode"`
	CurrencySymbol      string  `json:"currencySymbol"`
}

type Summary struct {
	Price         Price    `json:"price"`
	PreviousPrice *float64 `json:"previousPrice"` // Nullable
	NewRoute      bool     `json:"newRoute"`
}

type ExchangeRates struct {
	Table    string `json:"table"`
	Currency string `json:"currency"`
	Code     string `json:"code"`
	Rates    []Rate `json:"rates"`
}

type Rate struct {
	No            string  `json:"no"`
	EffectiveDate string  `json:"effectiveDate"`
	Bid           float64 `json:"bid"`
	Ask           float64 `json:"ask"`
}

const (
	lookForwardInMonths = 6
	chopinAirportCode   = "WAW"
	modlinAirportCode   = "WMI"
	alicanteAirportCode = "ALC"
)

func main() {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	startDate := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	endDate := startDate.AddDate(0, lookForwardInMonths, -1)

	var modlinToAlicante FlightResponse
	var chopinToAlicante FlightResponse
	var alicanteToModlin FlightResponse
	var alicanteToChopin FlightResponse
	var flightsData []byte
	flightsData = getRyanFlights(modlinAirportCode, alicanteAirportCode, startDate, endDate)
	err := json.Unmarshal([]byte(flightsData), &modlinToAlicante)
	if err != nil {
		log.Fatal("Error unmarshalling JSON:", err)
	}
	flightsData = getRyanFlights(chopinAirportCode, alicanteAirportCode, startDate, endDate)
	err = json.Unmarshal([]byte(flightsData), &chopinToAlicante)
	if err != nil {
		log.Fatal("Error unmarshalling JSON:", err)
	}
	flightsData = getRyanFlights(alicanteAirportCode, modlinAirportCode, startDate, endDate)
	err = json.Unmarshal([]byte(flightsData), &alicanteToModlin)
	if err != nil {
		log.Fatal("Error unmarshalling JSON:", err)
	}
	flightsData = getRyanFlights(alicanteAirportCode, chopinAirportCode, startDate, endDate)
	err = json.Unmarshal([]byte(flightsData), &alicanteToChopin)
	if err != nil {
		log.Fatal("Error unmarshalling JSON:", err)
	}

	warsawToAlicanteFares := append(modlinToAlicante.Fares, chopinToAlicante.Fares...)
	alicanteToWarsawFares := append(alicanteToModlin.Fares, alicanteToChopin.Fares...)
	convertEURtoPLN(&alicanteToWarsawFares)

	for _, fare := range alicanteToWarsawFares {
		fmt.Printf("Flight from %s (%s) to %s (%s), price: %.2f %s, curr symbol: %s, date: %s\n",
			fare.Outbound.DepartureAirport.Name,
			fare.Outbound.DepartureAirport.IATACode,
			fare.Outbound.ArrivalAirport.Name,
			fare.Outbound.ArrivalAirport.IATACode,
			fare.Outbound.Price.Value,
			fare.Outbound.Price.CurrencyCode,
			fare.Outbound.Price.CurrencySymbol,
			fare.Outbound.DepartureDate)
	}

	for _, wawToAlc := range warsawToAlicanteFares {
		for _, alcToWaw := range alicanteToWarsawFares {
			departureDate, err := time.Parse("2006-01-02T15:04:05", wawToAlc.Outbound.DepartureDate)
			if err != nil {
				log.Fatal(err)
			}
			returnDate, err := time.Parse("2006-01-02T15:04:05", alcToWaw.Outbound.DepartureDate)
			if err != nil {
				log.Fatal(err)
			}
			if departureDate.Before(returnDate) && returnDate.Sub(departureDate) < time.Hour*24*15 {
				log.Print(wawToAlc.Outbound.DepartureDate, "---", alcToWaw.Outbound.DepartureDate)
				log.Print(wawToAlc.Outbound.Price.Value + alcToWaw.Outbound.Price.Value)
			}
		}
	}

	// if 2 == 2 {
	// 	log.Print()
	// }

}

func getRyanFlights(
	departureAirportCode string,
	arrivalAirportCode string,
	startDate time.Time,
	endDate time.Time,
) []byte {
	url := fmt.Sprintf(
		"https://www.ryanair.com/api/farfnd/v4/oneWayFares?departureAirportIataCode=%s&outboundDepartureDateFrom=%s&market=pl-pl&adultPaxCount=1&arrivalAirportIataCode=%s&searchMode=ALL&outboundDepartureDateTo=%s&outboundDepartureDaysOfWeek=MONDAY,TUESDAY,WEDNESDAY,THURSDAY,FRIDAY,SATURDAY,SUNDAY&outboundDepartureTimeFrom=00:00&outboundDepartureTimeTo=23:59",
		departureAirportCode,
		startDate.Format(time.DateOnly),
		arrivalAirportCode,
		endDate.Format(time.DateOnly),
	)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: received non-200 response code: %d\n", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading the response body:", err)
	}

	return body
}

func convertEURtoPLN(fares *[]Fare) {
	euroRate := getEuroRate()
	for i := range *fares {
		(*fares)[i].Outbound.Price.Value *= euroRate
		(*fares)[i].Outbound.Price.CurrencyCode = "PLN"
		(*fares)[i].Outbound.Price.CurrencySymbol = "zł"
	}
}

func getEuroRate() float64 {
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://api.nbp.pl/api/exchangerates/rates/c/eur/today/?format=json", bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11`)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: received non-200 response code: %d\n", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading the response body:", err)
	}

	var exchangeRates ExchangeRates
	err = json.Unmarshal([]byte(body), &exchangeRates)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v\n", err)
	}

	return exchangeRates.Rates[0].Ask
}
