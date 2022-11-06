package helpers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"microservice/structs"
)

// RequestHasQueryParameters returns a byte which shows the parameters available in the request. The single bits are
// ordered by the parameter array supplied to this function
func RequestHasQueryParameters(parameters []string, request *http.Request) bool {
	var returnValue = false
	for _, parameter := range parameters {
		if request.URL.Query().Has(parameter) {
			returnValue = true
		} else {
			returnValue = false
		}
	}
	return returnValue
}

// ReadPrognosisResultFile returns the results start year, end year, lower bound, medium bound, upper bound
func ReadPrognosisResultFile(fileName string) []structs.ResultDataPoint {
	var dataPoints []structs.ResultDataPoint

	// Try to read the file
	fileContents, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil
	}

	jsonError := json.Unmarshal(fileContents, &dataPoints)
	if jsonError != nil {
		return nil
	}

	return dataPoints
}
