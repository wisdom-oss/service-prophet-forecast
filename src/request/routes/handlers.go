// Package routes
// This package contains all route handlers for the microservice
package routes

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gosimple/slug"
	"github.com/lib/pq"
	"microservice/request/enums"
	requestErrors "microservice/request/error"
	"microservice/structs"
	"microservice/utils"
	"microservice/vars"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

/*
ForecastRequest

This handler shows how a basic handler works and how to send back a message
*/
func ForecastRequest(responseWriter http.ResponseWriter, request *http.Request) {
	// get the shape keys that are set in the query url
	ctxShapeKeys := request.Context().Value("key")

	// check if any keys have been set
	if ctxShapeKeys == nil {
		// build a request error and send it back
		requestError, err := requestErrors.BuildRequestError(requestErrors.MissingShapeKeys)
		if err != nil {
			requestErrors.RespondWithInternalError(err, responseWriter)
			return
		}
		requestErrors.RespondWithRequestError(requestError, responseWriter)
		return
	}

	// since we have shape keys they will now be put into a string array
	shapeKeys := ctxShapeKeys.([]string)

	// now build a regex which matches any key and their possible children in the database
	shapeKeyRegEx := "("
	for _, shapeKey := range shapeKeys {
		if len(shapeKey) < 12 {
			missingNums := 12 - len(shapeKey)
			shapeKeyRegEx += fmt.Sprintf(`%s\d{%d}|`, shapeKey, missingNums)
		} else {
			shapeKeyRegEx += fmt.Sprintf(`%s|`, shapeKey)
		}
	}
	shapeKeyRegEx = strings.Trim(shapeKeyRegEx, "|")
	shapeKeyRegEx += ")"

	vars.HttpLogger.Info().Msg("getting municipality keys")
	// now query the database for the municipal keys matching the query
	shapeKeyRows, queryError := vars.SqlQueries.Query(vars.PostgresConnection, "get-full-municipality-keys", shapeKeyRegEx)
	if queryError != nil {
		// send back an error response
		requestErrors.RespondWithInternalError(queryError, responseWriter)
		return
	}

	// now iterate through the query response and put the municipality keys into an array
	var municipalityKeys []string

	for shapeKeyRows.Next() {
		var municipalityKey string
		scanError := shapeKeyRows.Scan(&municipalityKey)

		if scanError != nil {
			requestErrors.RespondWithInternalError(scanError, responseWriter)
			return
		}

		municipalityKeys = append(municipalityKeys, municipalityKey)
	}

	// now prepare to get the water usage data from the database
	vars.HttpLogger.Info().Msg("pulling water usage data")
	waterUsageRows, queryError := vars.SqlQueries.Query(vars.PostgresConnection, "get-water-usages", pq.Array(municipalityKeys))
	if queryError != nil {
		// send back an error response
		requestErrors.RespondWithInternalError(queryError, responseWriter)
		return
	}
	waterUsageData, err := utils.ReadDataForProphet(waterUsageRows)
	if err != nil {
		requestErrors.RespondWithInternalError(err, responseWriter)
		return
	}

	if len(waterUsageData) == 0 {
		// no water usage records have been found. send an error
		requestError, err := requestErrors.BuildRequestError(requestErrors.NoWaterUsageData)
		if err != nil {
			requestErrors.RespondWithInternalError(err, responseWriter)
			return
		}
		requestErrors.RespondWithRequestError(requestError, responseWriter)
		return
	}

	// now determine the first year of the water usage data to determine the first year of population data needed
	datasetStartYear := strings.Split(waterUsageData[0].Date, "-")[0]

	// now get the current population data from the database
	vars.HttpLogger.Info().Msg("pulling current population data")
	currentPopulationRows, queryError := vars.SqlQueries.Query(vars.PostgresConnection, "get-current-population",
		pq.Array(municipalityKeys),
		datasetStartYear)
	if queryError != nil {
		// send back an error response
		requestErrors.RespondWithInternalError(queryError, responseWriter)
		return
	}
	currentPopulationData, err := utils.ReadDataForProphet(currentPopulationRows)
	if err != nil {
		requestErrors.RespondWithInternalError(err, responseWriter)
		return
	}

	// now get the predicted population data from the database
	vars.HttpLogger.Info().Msg("pulling low migration population data")
	lowMigrationLevelPopulationRows, queryError := vars.SqlQueries.Query(vars.PostgresConnection,
		"get-future-population", pq.Array(municipalityKeys), enums.LowMigrationLevel)
	if queryError != nil {
		// send back an error response
		requestErrors.RespondWithInternalError(queryError, responseWriter)
		return
	}
	lowPopulationMigrationData, err := utils.ReadDataForProphet(lowMigrationLevelPopulationRows)
	if err != nil {
		requestErrors.RespondWithInternalError(err, responseWriter)
		return
	}

	vars.HttpLogger.Info().Msg("pulling medium migration population data")
	mediumMigrationLevelPopulationRows, queryError := vars.SqlQueries.Query(vars.PostgresConnection,
		"get-future-population", pq.Array(municipalityKeys), enums.MediumMigrationLevel)
	if queryError != nil {
		// send back an error response
		requestErrors.RespondWithInternalError(queryError, responseWriter)
		return
	}
	mediumPopulationMigrationData, err := utils.ReadDataForProphet(mediumMigrationLevelPopulationRows)
	if err != nil {
		requestErrors.RespondWithInternalError(err, responseWriter)
		return
	}

	vars.HttpLogger.Info().Msg("pulling high migration population data")
	highMigrationLevelPopulationRows, queryError := vars.SqlQueries.Query(vars.PostgresConnection,
		"get-future-population", pq.Array(municipalityKeys), enums.HighMigrationLevel)
	if queryError != nil {
		// send back an error response
		requestErrors.RespondWithInternalError(queryError, responseWriter)
		return
	}
	highPopulationMigrationData, err := utils.ReadDataForProphet(highMigrationLevelPopulationRows)
	if err != nil {
		requestErrors.RespondWithInternalError(err, responseWriter)
		return
	}

	// prepare the file names by making a slug from the generated request id from the context
	slugRequestID := slug.Make(middleware.GetReqID(request.Context()))

	currentPopulationDataFileName := fmt.Sprintf("current_population_%s.json", slugRequestID)
	lowMigrationDataFileName := fmt.Sprintf("low_population_migration_%s.json", slugRequestID)
	mediumMigrationDataFileName := fmt.Sprintf("medium_population_migration_%s.json", slugRequestID)
	highMigrationDataFileName := fmt.Sprintf("high_population_migration_%s.json", slugRequestID)
	waterUsageDataFileName := fmt.Sprintf("water_usage_%s.json", slugRequestID)

	// write the data from the objects into the json files
	vars.HttpLogger.Info().Msg("writing pulled data to files")
	_, err = utils.WriteDataToFile(currentPopulationData, currentPopulationDataFileName)
	if err != nil {
		requestErrors.RespondWithInternalError(err, responseWriter)
		return
	}
	_, err = utils.WriteDataToFile(lowPopulationMigrationData, lowMigrationDataFileName)
	if err != nil {
		requestErrors.RespondWithInternalError(err, responseWriter)
		return
	}
	_, err = utils.WriteDataToFile(mediumPopulationMigrationData, mediumMigrationDataFileName)
	if err != nil {
		requestErrors.RespondWithInternalError(err, responseWriter)
		return
	}
	_, err = utils.WriteDataToFile(highPopulationMigrationData, highMigrationDataFileName)
	if err != nil {
		requestErrors.RespondWithInternalError(err, responseWriter)
		return
	}
	_, err = utils.WriteDataToFile(waterUsageData, waterUsageDataFileName)
	if err != nil {
		requestErrors.RespondWithInternalError(err, responseWriter)
		return
	}

	// now execute the r script from the res folder
	Rscript := exec.Command("Rscript", "./res/prophet.r", slugRequestID, vars.TemporaryDataDirectory)
	Rscript.Stdout = os.Stdout
	vars.HttpLogger.Info().Msg("starting prognosis via rscript")
	executionStartTime := time.Now()
	err = Rscript.Run()
	if err != nil {
		requestErrors.RespondWithInternalError(err, responseWriter)
		return
	}
	executionTime := time.Since(executionStartTime)
	vars.HttpLogger.Info().Str("executionTime", executionTime.String()).Msg("finished prognosis via rscript")

	// now load the result files
	lowMigrationResultFileName := fmt.Sprintf("result_low_migration_%s.json", slugRequestID)
	mediumMigrationResultFileName := fmt.Sprintf("result_medium_migration_%s.json", slugRequestID)
	highMigrationResultFileName := fmt.Sprintf("result_high_migration_%s.json", slugRequestID)

	lowMigrationForecast := utils.ReadPrognosisResultFile(lowMigrationResultFileName)
	mediumMigrationForecast := utils.ReadPrognosisResultFile(mediumMigrationResultFileName)
	highMigrationForecast := utils.ReadPrognosisResultFile(highMigrationResultFileName)

	// now build the response and send it back
	response := structs.Response{
		LowMigrationData:    lowMigrationForecast,
		MediumMigrationData: mediumMigrationForecast,
		HighMigrationData:   highMigrationForecast,
	}

	responseWriter.Header().Set("Content-Type", "text/json")
	encodingError := json.NewEncoder(responseWriter).Encode(response)
	if encodingError != nil {
		requestErrors.RespondWithInternalError(encodingError, responseWriter)
		return
	}

}
