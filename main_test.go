package main

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func Test_convertEURtoPLN(t *testing.T) {
	mockFares1 := getMockAlcToWawFares()
	mockFares2 := getMockAlcToWawFares()
	mockFares3 := getMockAlcToWawFares()
	type args struct {
		fares    *[]Fare
		euroRate float64
	}
	tests := []struct {
		name string
		args args
		want []Fare
	}{
		{
			name: "test convert eur to pln 4.35 rate",
			args: args{&mockFares1, 4.35},
			want: getWantedAlcToWawFares_1(),
		},
		{
			name: "test convert eur to pln 4.00 rate",
			args: args{&mockFares2, 4.00},
			want: getWantedAlcToWawFares_2(),
		},
		{
			name: "test convert eur to pln 3.87 rate",
			args: args{&mockFares3, 3.87},
			want: getWantedAlcToWawFares_3(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			convertEURtoPLN(test.args.fares, test.args.euroRate)
			if !cmp.Equal(*test.args.fares, test.want) {
				t.Errorf("\n%v\n!=\n%v", *test.args.fares, test.want)
			}
		})
	}
}

func Test_getFlightsToCompare(t *testing.T) {
	type args struct {
		wawToAlcDates []string
		alcToWawDates []string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "find 4 cross pairs of flights to compare",
			args: args{
				[]string{"2024-10-05T19:15:00", "2024-10-06T11:25:00", "2024-11-11T06:25:00", "2024-11-12T06:25:00"},
				[]string{"2024-10-12T19:15:00", "2024-10-10T11:25:00", "2024-11-16T06:25:00", "2024-11-18T06:25:00"},
			},
			want: 8,
		},
		{
			name: "find 8 cross pairs of flights to compare",
			args: args{
				[]string{"2024-10-05T19:15:00", "2024-10-06T11:25:00", "2024-10-07T06:25:00", "2024-10-08T06:25:00"},
				[]string{"2024-10-12T19:15:00", "2024-10-13T11:25:00", "2024-10-14T06:25:00", "2024-10-15T06:25:00"},
			},
			want: 16,
		},
		{
			name: "check transition between deadlines",
			args: args{
				[]string{"2024-10-25T19:15:00", "2024-12-02T11:25:00", "2024-12-02T06:25:00", "2024-11-12T06:25:00"},
				[]string{"2024-10-28T19:15:00", "2024-12-17T11:25:00", "2024-12-01T06:25:00", "2025-11-16T06:25:00"},
			},
			want: 0,
		},
		{
			name: "find 2 flights to compare",
			args: args{
				[]string{"2024-10-29T19:15:00", "2024-12-01T11:25:00", "2024-12-02T06:25:00", "2024-04-02T06:25:00"},
				[]string{"2024-10-29T19:15:00", "2024-12-10T11:25:00", "2024-12-02T06:25:00", "2024-12-02T06:25:00"},
			},
			want: 2,
		},
		{
			name: "find 3 flights to compare",
			args: args{
				[]string{"2024-10-25T19:15:00", "2024-12-01T11:25:00", "2024-12-02T06:25:00", "2024-11-12T06:25:00"},
				[]string{"2024-10-29T19:15:00", "2024-12-10T11:25:00", "2024-12-02T06:25:00", "2024-11-14T06:25:00"},
			},
			want: 3,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flights, err := getFlightsToCompare(
				getMockWawToAlcFaresFlexDates(test.args.wawToAlcDates...),
				getMockAlcToWawFaresFlexDates(test.args.alcToWawDates...),
			)
			if err != nil {
				t.Error(err)
			}
			keys := []time.Month{}
			got := 0
			for key := range flights {
				keys = append(keys, key)
			}
			for _, key := range keys {
				got += len(flights[key])
			}
			if got != test.want {
				t.Errorf("flights to compare length, got: %d != want: %d", got, test.want)
			}
		})
	}
}

func getMockAlcToWawFares() []Fare {
	return []Fare{
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-10-29T19:15:00",
				ArrivalDate:   "2024-10-29T22:55:00",
				Price: Price{
					Value:               95.00,
					ValueMainUnit:       "95",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				FlightKey:     "FR~1001~ ~~WMI~10/29/2024 19:15~ALC~10/29/2024 22:55~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729187269000,
			},
			Summary: Summary{
				Price: Price{
					Value:               95.00,
					ValueMainUnit:       "95",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-12-10T11:25:00",
				ArrivalDate:   "2024-12-10T15:05:00",
				Price: Price{
					Value:               119.00,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/01/2024 11:25~ALC~12/01/2024 15:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729169125000,
			},
			Summary: Summary{
				Price: Price{
					Value:               119.00,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-12-02T06:25:00",
				ArrivalDate:   "2024-12-02T10:05:00",
				Price: Price{
					Value:               1,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/02/2024 06:25~ALC~12/02/2024 10:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729168899000,
			},
			Summary: Summary{
				Price: Price{
					Value:               1,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-12-02T06:25:00",
				ArrivalDate:   "2024-12-02T10:05:00",
				Price: Price{
					Value:               1099,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/02/2024 06:25~ALC~12/02/2024 10:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729168899000,
			},
			Summary: Summary{
				Price: Price{
					Value:               1099,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
	}
}

func getMockWawToAlcFaresFlexDates(dates ...string) []Fare {
	return []Fare{
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				DepartureDate: dates[0],
				ArrivalDate:   "2024-10-29T22:55:00",
				Price: Price{
					Value:               95.00,
					ValueMainUnit:       "95",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~10/29/2024 19:15~ALC~10/29/2024 22:55~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729187269000,
			},
			Summary: Summary{
				Price: Price{
					Value:               95.00,
					ValueMainUnit:       "95",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				DepartureDate: dates[1],
				ArrivalDate:   "2024-12-01T15:05:00",
				Price: Price{
					Value:               119.00,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/01/2024 11:25~ALC~12/01/2024 15:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729169125000,
			},
			Summary: Summary{
				Price: Price{
					Value:               119.00,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				DepartureDate: dates[2],
				ArrivalDate:   "2024-12-02T10:05:00",
				Price: Price{
					Value:               1,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/02/2024 06:25~ALC~12/02/2024 10:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729168899000,
			},
			Summary: Summary{
				Price: Price{
					Value:               1,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				DepartureDate: dates[3],
				ArrivalDate:   "2024-04-02T10:05:00",
				Price: Price{
					Value:               1099,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/02/2024 06:25~ALC~12/02/2024 10:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729168899000,
			},
			Summary: Summary{
				Price: Price{
					Value:               1099,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
	}
}

func getMockAlcToWawFaresFlexDates(dates ...string) []Fare {
	return []Fare{
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: dates[0],
				ArrivalDate:   "2024-10-29T22:55:00",
				Price: Price{
					Value:               95.00,
					ValueMainUnit:       "95",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				FlightKey:     "FR~1001~ ~~WMI~10/29/2024 19:15~ALC~10/29/2024 22:55~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729187269000,
			},
			Summary: Summary{
				Price: Price{
					Value:               95.00,
					ValueMainUnit:       "95",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: dates[1],
				ArrivalDate:   "2024-12-10T15:05:00",
				Price: Price{
					Value:               119.00,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/01/2024 11:25~ALC~12/01/2024 15:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729169125000,
			},
			Summary: Summary{
				Price: Price{
					Value:               119.00,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: dates[2],
				ArrivalDate:   "2024-12-02T10:05:00",
				Price: Price{
					Value:               1,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/02/2024 06:25~ALC~12/02/2024 10:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729168899000,
			},
			Summary: Summary{
				Price: Price{
					Value:               1,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: dates[3],
				ArrivalDate:   "2024-12-02T10:05:00",
				Price: Price{
					Value:               1099,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/02/2024 06:25~ALC~12/02/2024 10:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729168899000,
			},
			Summary: Summary{
				Price: Price{
					Value:               1099,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "EUR",
					CurrencySymbol:      "€",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
	}
}

func getWantedAlcToWawFares_1() []Fare {
	return []Fare{
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-10-29T19:15:00",
				ArrivalDate:   "2024-10-29T22:55:00",
				Price: Price{
					Value:               413.25,
					ValueMainUnit:       "95",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~10/29/2024 19:15~ALC~10/29/2024 22:55~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729187269000,
			},
			Summary: Summary{
				Price: Price{
					Value:               413.25,
					ValueMainUnit:       "95",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-12-10T11:25:00",
				ArrivalDate:   "2024-12-10T15:05:00",
				Price: Price{
					Value:               517.65,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/01/2024 11:25~ALC~12/01/2024 15:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729169125000,
			},
			Summary: Summary{
				Price: Price{
					Value:               517.65,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-12-02T06:25:00",
				ArrivalDate:   "2024-12-02T10:05:00",
				Price: Price{
					Value:               4.35,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/02/2024 06:25~ALC~12/02/2024 10:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729168899000,
			},
			Summary: Summary{
				Price: Price{
					Value:               4.35,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-12-02T06:25:00",
				ArrivalDate:   "2024-12-02T10:05:00",
				Price: Price{
					Value:               4780.65,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/02/2024 06:25~ALC~12/02/2024 10:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729168899000,
			},
			Summary: Summary{
				Price: Price{
					Value:               4780.65,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
	}
}
func getWantedAlcToWawFares_2() []Fare {
	return []Fare{
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-10-29T19:15:00",
				ArrivalDate:   "2024-10-29T22:55:00",
				Price: Price{
					Value:               380.00,
					ValueMainUnit:       "95",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~10/29/2024 19:15~ALC~10/29/2024 22:55~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729187269000,
			},
			Summary: Summary{
				Price: Price{
					Value:               380.00,
					ValueMainUnit:       "95",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-12-10T11:25:00",
				ArrivalDate:   "2024-12-10T15:05:00",
				Price: Price{
					Value:               476.00,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/01/2024 11:25~ALC~12/01/2024 15:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729169125000,
			},
			Summary: Summary{
				Price: Price{
					Value:               476.00,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-12-02T06:25:00",
				ArrivalDate:   "2024-12-02T10:05:00",
				Price: Price{
					Value:               4.00,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/02/2024 06:25~ALC~12/02/2024 10:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729168899000,
			},
			Summary: Summary{
				Price: Price{
					Value:               4.00,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-12-02T06:25:00",
				ArrivalDate:   "2024-12-02T10:05:00",
				Price: Price{
					Value:               4396.00,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/02/2024 06:25~ALC~12/02/2024 10:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729168899000,
			},
			Summary: Summary{
				Price: Price{
					Value:               4396.00,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
	}
}
func getWantedAlcToWawFares_3() []Fare {
	return []Fare{
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-10-29T19:15:00",
				ArrivalDate:   "2024-10-29T22:55:00",
				Price: Price{
					Value:               367.65,
					ValueMainUnit:       "95",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~10/29/2024 19:15~ALC~10/29/2024 22:55~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729187269000,
			},
			Summary: Summary{
				Price: Price{
					Value:               367.65,
					ValueMainUnit:       "95",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-12-10T11:25:00",
				ArrivalDate:   "2024-12-10T15:05:00",
				Price: Price{
					Value:               460.53,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/01/2024 11:25~ALC~12/01/2024 15:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729169125000,
			},
			Summary: Summary{
				Price: Price{
					Value:               460.53,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-12-02T06:25:00",
				ArrivalDate:   "2024-12-02T10:05:00",
				Price: Price{
					Value:               3.87,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/02/2024 06:25~ALC~12/02/2024 10:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729168899000,
			},
			Summary: Summary{
				Price: Price{
					Value:               3.87,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
		{
			Outbound: Outbound{
				DepartureAirport: Airport{
					CountryName: "Hiszpania",
					IATACode:    "ALC",
					Name:        "Alicante",
					SEOName:     "alicante",
					City: City{
						Name:        "Alicante",
						Code:        "ALICANTE",
						CountryCode: "es",
					},
				},
				ArrivalAirport: Airport{
					CountryName: "Polska",
					IATACode:    "WMI",
					Name:        "Warszawa-Modlin",
					SEOName:     "warsaw-modlin",
					City: City{
						Name:        "Warszawa",
						Code:        "WARSAW",
						MacCode:     "WWA",
						CountryCode: "pl",
					},
				},
				DepartureDate: "2024-12-02T06:25:00",
				ArrivalDate:   "2024-12-02T10:05:00",
				Price: Price{
					Value:               4253.13,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				FlightKey:     "FR~1001~ ~~WMI~12/02/2024 06:25~ALC~12/02/2024 10:05~~",
				FlightNumber:  "FR1001",
				PreviousPrice: nil,
				PriceUpdated:  1729168899000,
			},
			Summary: Summary{
				Price: Price{
					Value:               4253.13,
					ValueMainUnit:       "119",
					ValueFractionalUnit: "00",
					CurrencyCode:        "PLN",
					CurrencySymbol:      "zł",
				},
				PreviousPrice: nil,
				NewRoute:      false,
			},
		},
	}
}
