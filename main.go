package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	lookForwardInMonths = 5
	offersPerMonth      = 5
	chopinAirportCode   = "WAW"
	modlinAirportCode   = "WMI"
	alicanteAirportCode = "ALC"
)

var (
	minTripDurationInDays = 3
	maxTripDurationInDays = 15
	chatId, botToken      string
)

func main() {
	if err := setOsArgs(); err != nil {
		log.Fatal(err)
	}
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	startDate := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	endDate := startDate.AddDate(0, lookForwardInMonths, -1)

	warsawToAlicanteFares, err := getWarsawToAlicanteFlights(startDate, endDate)
	if err != nil {
		log.Fatal(err)
	}
	alicanteToWarsawFares, err := getAlicanteToWarsawFlights(startDate, endDate)
	if err != nil {
		log.Fatal(err)
	}
	euroRate, err := getEuroRate()
	if err != nil {
		log.Fatal(err)
	}
	convertEURtoPLN(&alicanteToWarsawFares, euroRate)

	flightsToCompare, err := getFlightsToCompare(warsawToAlicanteFares, alicanteToWarsawFares)
	if err != nil {
		log.Fatal(err)
	}
	message := buildMessage(now, flightsToCompare)
	sendMessageToTelegram(message, botToken, chatId)
}

func setOsArgs() error {
	if len(os.Args) <= 1 {
		return errors.New("missing required arguments: chatId, botToken.\nadditional arguments are: minTripDurationInDays and maxTripDurationInDays")
	} else {
		if len(os.Args) > 4 {
			min, err := strconv.Atoi(os.Args[3])
			if err != nil {
				return err
			}
			max, err := strconv.Atoi(os.Args[4])
			if err != nil {
				return err
			}
			if min <= 0 || max <= 0 {
				return fmt.Errorf("one of optional args is wrong. integer greater than 0 needed. your args: %s, %s", os.Args[3], os.Args[4])
			}
			minTripDurationInDays = min
			maxTripDurationInDays = max
		}
		chatId, botToken = os.Args[1], os.Args[2]
	}
	return nil
}

func getWarsawToAlicanteFlights(startDate, endDate time.Time) ([]Fare, error) {
	var modlinToAlicante FlightResponse
	var chopinToAlicante FlightResponse
	var flightsData []byte

	flightsData, err := getRyanFlights(modlinAirportCode, alicanteAirportCode, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("could not gather data from ryanair website.\n%v", err)
	}
	err = json.Unmarshal([]byte(flightsData), &modlinToAlicante)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON response: %v", err)
	}
	flightsData, err = getRyanFlights(chopinAirportCode, alicanteAirportCode, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("could not gather data from ryanair website.\n%v", err)
	}
	err = json.Unmarshal([]byte(flightsData), &chopinToAlicante)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON response: %v", err)
	}
	return append(modlinToAlicante.Fares, chopinToAlicante.Fares...), nil
}

func getAlicanteToWarsawFlights(startDate, endDate time.Time) ([]Fare, error) {
	var alicanteToModlin FlightResponse
	var alicanteToChopin FlightResponse
	var flightsData []byte
	flightsData, err := getRyanFlights(alicanteAirportCode, modlinAirportCode, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("could not gather data from ryanair website.\n%v", err)
	}
	err = json.Unmarshal([]byte(flightsData), &alicanteToModlin)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON response: %v", err)
	}
	flightsData, err = getRyanFlights(alicanteAirportCode, chopinAirportCode, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("could not gather data from ryanair website.\n%v", err)
	}
	err = json.Unmarshal([]byte(flightsData), &alicanteToChopin)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON response: %v", err)
	}
	return append(alicanteToModlin.Fares, alicanteToChopin.Fares...), nil
}

func getRyanFlights(
	departureAirportCode string,
	arrivalAirportCode string,
	startDate time.Time,
	endDate time.Time,
) ([]byte, error) {
	url := fmt.Sprintf(
		"https://www.ryanair.com/api/farfnd/v4/oneWayFares?departureAirportIataCode=%s&outboundDepartureDateFrom=%s&market=pl-pl&adultPaxCount=1&arrivalAirportIataCode=%s&searchMode=ALL&outboundDepartureDateTo=%s&outboundDepartureDaysOfWeek=MONDAY,TUESDAY,WEDNESDAY,THURSDAY,FRIDAY,SATURDAY,SUNDAY&outboundDepartureTimeFrom=00:00&outboundDepartureTimeTo=23:59",
		departureAirportCode,
		startDate.Format(time.DateOnly),
		arrivalAirportCode,
		endDate.Format(time.DateOnly),
	)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: received non-200 response code: %d\nurl: %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading the response body: %v", err)
	}

	return body, nil
}

func convertEURtoPLN(fares *[]Fare, euroRate float64) {
	for i := range *fares {
		convertedValue := (*fares)[i].Outbound.Price.Value * euroRate
		(*fares)[i].Outbound.Price.Value = math.Round(convertedValue*100) / 100
		(*fares)[i].Summary.Price.Value = math.Round(convertedValue*100) / 100
		(*fares)[i].Outbound.Price.CurrencyCode = "PLN"
		(*fares)[i].Summary.Price.CurrencyCode = "PLN"
		(*fares)[i].Outbound.Price.CurrencySymbol = "zł"
		(*fares)[i].Summary.Price.CurrencySymbol = "zł"
	}
}

func getEuroRate() (float64, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://api.nbp.pl/api/exchangerates/rates/a/eur/last/1/?format=json", bytes.NewBuffer([]byte{}))
	if err != nil {
		return 0, err
	}
	req.Header.Add("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11`)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("getEuroRate() Error: received non-200 response code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading the response body: %v", err)
	}

	var exchangeRates ExchangeRates
	err = json.Unmarshal([]byte(body), &exchangeRates)
	if err != nil {
		return 0, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return exchangeRates.Rates[0].Mid, nil
}

func getFlightsToCompare(warsawToAlicanteFares, alicanteToWarsawFares []Fare) (map[time.Month][]FlightToCompare, error) {
	flights := make(map[time.Month][]FlightToCompare)
	for _, wawToAlc := range warsawToAlicanteFares {
		for _, alcToWaw := range alicanteToWarsawFares {
			departureDate, err := time.Parse("2006-01-02T15:04:05", wawToAlc.Outbound.DepartureDate)
			if err != nil {
				return nil, err
			}
			returnDate, err := time.Parse("2006-01-02T15:04:05", alcToWaw.Outbound.DepartureDate)
			if err != nil {
				return nil, err
			}
			if departureDate.Before(returnDate) &&
				returnDate.Sub(departureDate) < time.Hour*24*time.Duration(maxTripDurationInDays) &&
				returnDate.Sub(departureDate) > time.Hour*24*time.Duration(minTripDurationInDays) {

				flights[departureDate.Month()] = append(
					flights[departureDate.Month()], FlightToCompare{wawToAlc.Outbound, alcToWaw.Outbound})
			}
		}
	}
	return flights, nil
}

func buildMessage(now time.Time, flightsToCompare map[time.Month][]FlightToCompare) bytes.Buffer {
	upcomingMonths := make([]time.Month, lookForwardInMonths)
	for i := 0; i < lookForwardInMonths; i++ {
		upcomingMonths[i] = now.AddDate(0, i, -now.Day()+1).Month()
	}

	var message bytes.Buffer
	for _, month := range upcomingMonths {
		sort.Slice(flightsToCompare[month], func(i, j int) bool {
			priceSummary := flightsToCompare[month][i].AbroadFlight.Price.Value + flightsToCompare[month][i].ReturnFlight.Price.Value
			nextPriceSummary := flightsToCompare[month][j].AbroadFlight.Price.Value + flightsToCompare[month][j].ReturnFlight.Price.Value
			return priceSummary < nextPriceSummary
		})
		message.WriteString(month.String())
		message.WriteString("\n")
		if len(flightsToCompare[month]) > 0 {
			for _, trip := range flightsToCompare[month][:offersPerMonth] {
				message.WriteString(fmt.Sprintf("%s ---> %s ", trip.AbroadFlight.DepartureAirport.Name, trip.AbroadFlight.ArrivalAirport.Name))
				message.WriteString(fmt.Sprintf("%s ", strings.Replace(trip.AbroadFlight.DepartureDate, "T", " ", 1)))
				message.WriteString(fmt.Sprintf("%s%s\n", strconv.FormatFloat(trip.AbroadFlight.Price.Value, 'f', 2, 64), trip.AbroadFlight.Price.CurrencySymbol))
				message.WriteString(fmt.Sprintf("%s ---> %s ", trip.ReturnFlight.DepartureAirport.Name, trip.ReturnFlight.ArrivalAirport.Name))
				message.WriteString(fmt.Sprintf("%s ", strings.Replace(trip.ReturnFlight.DepartureDate, "T", " ", 1)))
				message.WriteString(fmt.Sprintf("%s%s\n", strconv.FormatFloat(trip.ReturnFlight.Price.Value, 'f', 2, 64), trip.ReturnFlight.Price.CurrencySymbol))
				message.WriteString(fmt.Sprintf("Razem: %szł\n", strconv.FormatFloat(trip.AbroadFlight.Price.Value+trip.ReturnFlight.Price.Value, 'f', 2, 64)))
				message.WriteString("\n")
			}
		} else {
			message.WriteString("No flights for this month\n")
		}
		message.WriteString("------------------------------------------\n")
	}
	return message
}

func sendMessageToTelegram(message bytes.Buffer, botToken, chatId string) error {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	data := url.Values{}
	data.Set("chat_id", chatId)
	data.Set("text", message.String())
	req, err := http.NewRequest("POST", u, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Print("Message sent successfully!")
	} else {
		return fmt.Errorf("failed to send message. Status code: %d", resp.StatusCode)
	}
	return nil
}
