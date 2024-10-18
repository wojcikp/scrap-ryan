package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

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

	chatId, botToken := os.Args[1], os.Args[2]

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

	var flightsToCompare []FlightToCompare
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
			if departureDate.Before(returnDate) &&
				returnDate.Sub(departureDate) < time.Hour*24*15 &&
				returnDate.Sub(departureDate) > time.Hour*24*3 {

				flightsToCompare = append(flightsToCompare, FlightToCompare{wawToAlc.Outbound, alcToWaw.Outbound})
			}
		}
	}

	sort.Slice(flightsToCompare, func(i, j int) bool {
		priceSummary := flightsToCompare[i].AbroadFlight.Price.Value + flightsToCompare[i].ReturnFlight.Price.Value
		nextPriceSummary := flightsToCompare[j].AbroadFlight.Price.Value + flightsToCompare[j].ReturnFlight.Price.Value
		return priceSummary < nextPriceSummary
	})

	var message bytes.Buffer
	for _, trip := range flightsToCompare[:10] {
		log.Print("|", trip.AbroadFlight.DepartureDate, "---", trip.ReturnFlight.DepartureDate)
		log.Print("|", trip.AbroadFlight.Price.Value+trip.ReturnFlight.Price.Value)
		log.Print("+-------------------------------------------+")
		message.WriteString(fmt.Sprintf("%s ---> %s ", trip.AbroadFlight.DepartureAirport.Name, trip.AbroadFlight.ArrivalAirport.Name))
		message.WriteString(fmt.Sprintf("%s ", strings.Replace(trip.AbroadFlight.DepartureDate, "T", " ", 1)))
		message.WriteString(fmt.Sprintf("%s%s\n", strconv.FormatFloat(trip.AbroadFlight.Price.Value, 'f', 2, 64), trip.AbroadFlight.Price.CurrencySymbol))
		message.WriteString(fmt.Sprintf("%s ---> %s ", trip.ReturnFlight.DepartureAirport.Name, trip.ReturnFlight.ArrivalAirport.Name))
		message.WriteString(fmt.Sprintf("%s ", strings.Replace(trip.ReturnFlight.DepartureDate, "T", " ", 1)))
		message.WriteString(fmt.Sprintf("%s%s\n", strconv.FormatFloat(trip.ReturnFlight.Price.Value, 'f', 2, 64), trip.ReturnFlight.Price.CurrencySymbol))
		message.WriteString(fmt.Sprintf("Razem: %szł\n", strconv.FormatFloat(trip.AbroadFlight.Price.Value+trip.ReturnFlight.Price.Value, 'f', 2, 64)))
		message.WriteString("\n")
	}

	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	data := url.Values{}
	data.Set("chat_id", chatId)
	data.Set("text", message.String())
	req, err := http.NewRequest("POST", u, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Print("Message sent successfully!")
	} else {
		log.Fatalf("Failed to send message. Status code: %d\n", resp.StatusCode)
	}

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
