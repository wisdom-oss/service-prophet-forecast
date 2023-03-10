package structs

// ScopeInformation contains the information about the scope for this service
type ScopeInformation struct {
	JSONSchema       string `json:"$schema"`
	ScopeName        string `json:"name"`
	ScopeDescription string `json:"description"`
	ScopeValue       string `json:"scopeStringValue"`
}

// RequestError contains all information about an error which shall be sent back to the client
type RequestError struct {
	HttpStatus       int    `json:"httpCode"`
	HttpError        string `json:"httpError"`
	ErrorCode        string `json:"error"`
	ErrorTitle       string `json:"errorName"`
	ErrorDescription string `json:"errorDescription"`
}

// InputDataPoint contains the water usage of a single year
type InputDataPoint struct {
	Date  string  `json:"ds"`
	Value float64 `json:"y"`
}

type OutputDataPoint struct {
	Date       string  `json:"ds"`
	LowerBound float64 `json:"lower"`
	Forecast   float64 `json:"forecast"`
	UpperBound float64 `json:"upper"`
}

type Response struct {
	LowMigrationData    []OutputDataPoint `json:"lowMigrationPrognosis"`
	MediumMigrationData []OutputDataPoint `json:"mediumMigrationPrognosis"`
	HighMigrationData   []OutputDataPoint `json:"highMigrationPrognosis"`
}
