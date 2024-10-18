package main

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

type FlightToCompare struct {
	AbroadFlight Outbound
	ReturnFlight Outbound
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
