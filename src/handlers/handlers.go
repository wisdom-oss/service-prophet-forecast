package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"

	"microservice/enums"
	e "microservice/errors"
	"microservice/helpers"
	"microservice/structs"
	"microservice/vars"

	"github.com/google/uuid"
)

func AuthorizationCheck(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(responseWriter http.ResponseWriter, request *http.Request) {
			logger := log.WithFields(
				log.Fields{
					"middleware": true,
					"title":      "AuthorizationCheck",
				},
			)
			logger.Debug("Checking the incoming request for authorization information set by the gateway")

			// Get the scopes the requesting user has
			scopes := request.Header.Get("X-Authenticated-Scope")
			// Check if the string is empty
			if strings.TrimSpace(scopes) == "" {
				logger.Warning("Unauthorized request detected. The required header had no content or was not set")
				helpers.SendRequestError(e.UnauthorizedRequest, responseWriter)
				return
			}

			scopeList := strings.Split(scopes, ",")
			if !helpers.StringArrayContains(scopeList, vars.ScopeConfiguration.ScopeValue) {
				logger.Error("Request rejected. The user is missing the scope needed for accessing this service")
				helpers.SendRequestError(e.MissingScope, responseWriter)
				return
			}
			// Call the next handler which will continue handling the request
			nextHandler.ServeHTTP(responseWriter, request)
		},
	)
}

/*
PingHandler

This handler is used to test if the service is able to ping itself. This is done to run a healthcheck on the container
*/
func PingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// ForecastRequestHandler reads the incoming request data and forwards it to the r script which than
// produces a response file
func ForecastRequestHandler(responseWriter http.ResponseWriter, request *http.Request) {
	logger := log.WithFields(
		log.Fields{
			"middleware": true,
			"title":      "ForecastRequestHandler",
		},
	)
	// Set the parameters which shall be parsed by this handler
	var parameters = []string{"key"}
	logger.WithField("request", request).Info("Got new request")
	// Check if the query parameters have been set
	availableParameters := helpers.RequestHasQueryParameters(parameters, request)
	switch availableParameters {
	case false:
		helpers.SendRequestError(e.MissingRequiredParameter, responseWriter)
		return
	case true:
		// Access the keys which have been sent
		shapeKeys := request.URL.Query()["key"]

		// Create a regular expression targeting all shapes in the database either starting with a key or matching
		// the key, if the key is 12 characters long
		var shapeKeyRegEx string
		for _, shapeKey := range shapeKeys {
			if len(shapeKey) < 12 {
				shapeKeyRegEx += fmt.Sprintf(`^%s\d+$`, shapeKey)
			} else {
				shapeKeyRegEx += fmt.Sprintf(`^%s$`, shapeKey)
			}
		}

		// Now build a query to get the current population data for the specified shapes
		currentPopulationQuery := `SELECT year, sum(population) as pop
								   FROM population.current
								   WHERE municipality_key ~ $1 AND year >= $2::int
								   GROUP BY year
								   ORDER BY year`

		// Now build a query to get the predicted population data for the specified shapes
		futurePopulationWithMigrationLevelQuery := `SELECT year, sum(population) as pop
													   FROM population.prognosis
													   WHERE municipal_key ~ $1 AND migration_level = $2::migration_level
													   GROUP BY year
													   ORDER BY year`

		// Now build a query to get the water usages only requesting those made by households (consumer group 2)
		waterUsageQuery := `SELECT year, sum(value) as usage
							FROM water_usage.usages
							WHERE municipal_key ~ $1
							AND consumer_group = 2
							GROUP BY year
							ORDER BY year`

		// Request the water usages
		logger.Info("Pulling water usage data from the database")
		waterUsageRows, queryError := vars.PostgresConnection.Query(waterUsageQuery, shapeKeyRegEx)
		if queryError != nil {
			logger.WithError(queryError).Error("An error occurred while getting the water usage data")
			helpers.SendRequestError(e.DatabaseQueryError, responseWriter)
			return
		}

		waterUsageData, err := helpers.ReadDataForProphet(waterUsageRows)
		if err != nil {
			logger.WithError(err).Error(
				"An error occurred while reading the population data from the returned" +
					"dataset",
			)
			helpers.SendRequestError(e.DatabaseQueryError, responseWriter)
		}

		// Now access the water usage data and check from where on the current population data shall be loaded
		firstWaterUsageDatasetYear := strings.Split(waterUsageData[0].Date, "-")[0]
		// Now get the population data from the database
		currentPopulationRows, queryError := vars.PostgresConnection.Query(
			currentPopulationQuery,
			shapeKeyRegEx,
			firstWaterUsageDatasetYear,
		)
		lowMigrationLevelPopulationRows, queryError := vars.PostgresConnection.Query(
			futurePopulationWithMigrationLevelQuery, shapeKeyRegEx, enums.LowMigrationLevel,
		)
		mediumMigrationLevelPopulationRows, queryError := vars.PostgresConnection.Query(
			futurePopulationWithMigrationLevelQuery, shapeKeyRegEx, enums.MediumMigrationLevel,
		)
		highMigrationLevelPopulationRows, queryError := vars.PostgresConnection.Query(
			futurePopulationWithMigrationLevelQuery, shapeKeyRegEx, enums.HighMigrationLevel,
		)
		if queryError != nil {
			logger.WithError(queryError).Error("An error occurred while getting the population data")
			helpers.SendRequestError(e.DatabaseQueryError, responseWriter)
			return
		}

		currentPopulationData, err := helpers.ReadDataForProphet(currentPopulationRows)
		lowMigrationLevelPopulationData, err := helpers.ReadDataForProphet(lowMigrationLevelPopulationRows)
		mediumMigrationLevelPopulationData, err := helpers.ReadDataForProphet(mediumMigrationLevelPopulationRows)
		highMigrationLevelPopulationData, err := helpers.ReadDataForProphet(highMigrationLevelPopulationRows)
		if err != nil {
			logger.WithError(err).Error(
				"An error occurred while reading the population data from the returned" +
					"dataset",
			)
			helpers.SendRequestError(e.DatabaseQueryError, responseWriter)
		}
		// Generate a new uuid for the request to identify the request later in the r script
		forecastUUID := uuid.New()

		// Create new temporary directory
		currentPopulationDataFileName := fmt.Sprintf("current_population_%s.json", forecastUUID.String())
		lowMigrationLevelPopulationDataFileName := fmt.Sprintf(
			"low_population_migration_%s.json",
			forecastUUID.String(),
		)
		mediumMigrationLevelPopulationDataFileName := fmt.Sprintf(
			"medium_population_migration_%s.json",
			forecastUUID.String(),
		)
		highMigrationLevelPopulationDataFileName := fmt.Sprintf(
			"high_population_migration_%s.json",
			forecastUUID.String(),
		)
		waterUsageDataFileName := fmt.Sprintf(
			"water_usage_%s.json",
			forecastUUID.String(),
		)

		// Now write the data from the database to the r script
		_, writeError := helpers.WriteDataToFile(currentPopulationData, currentPopulationDataFileName)
		if writeError != nil {
			helpers.SendRequestError(e.DataWriting, responseWriter)
			logger.WithError(writeError).Error(
				"An error occurred while writing the current population data to a file",
			)
			return
		}
		_, writeError = helpers.WriteDataToFile(
			lowMigrationLevelPopulationData,
			lowMigrationLevelPopulationDataFileName,
		)
		if writeError != nil {
			helpers.SendRequestError(e.DataWriting, responseWriter)
			logger.WithError(writeError).Error(
				"An error occurred while writing the low migration level population data to a file",
			)
			return
		}
		_, writeError = helpers.WriteDataToFile(
			mediumMigrationLevelPopulationData,
			mediumMigrationLevelPopulationDataFileName,
		)
		if writeError != nil {
			helpers.SendRequestError(e.DataWriting, responseWriter)
			logger.WithError(writeError).Error(
				"An error occurred while writing the medium migration level population data to a file",
			)
			return
		}
		_, writeError = helpers.WriteDataToFile(
			highMigrationLevelPopulationData,
			highMigrationLevelPopulationDataFileName,
		)
		if writeError != nil {
			helpers.SendRequestError(e.DataWriting, responseWriter)
			logger.WithError(writeError).Error(
				"An error occurred while writing the high migration level population data to a file",
			)
			return
		}
		_, writeError = helpers.WriteDataToFile(
			waterUsageData,
			waterUsageDataFileName,
		)
		if writeError != nil {
			helpers.SendRequestError(e.DataWriting, responseWriter)
			logger.WithError(writeError).Error(
				"An error occurred while writing the high migration level population data to a file",
			)
			return
		}

		// Now create a new command call for the rscript...
		rscriptCommand := exec.Command("Rscript", "./res/prophet.r", forecastUUID.String(), vars.TemporaryDataDirectory)
		rscriptCommand.Stdout = os.Stdout
		// ... and execute it
		logger.Info("Starting prognosis in RScript")
		executionErrors := rscriptCommand.Run()
		if executionErrors != nil {
			logger.WithError(executionErrors).Error("An error occurred while executing the r script")
			helpers.SendRequestError(e.RScriptError, responseWriter)
			return
		}
		logger.Info("Finished prognosis in RScript")

		// Now read the contents of the result files
		lowMigrationResultFileName := fmt.Sprintf(
			"result_low_migration_%s.json",
			forecastUUID.String(),
		)
		mediumMigrationResultFileName := fmt.Sprintf(
			"result_medium_migration_%s.json",
			forecastUUID.String(),
		)
		highMigrationResultFileName := fmt.Sprintf(
			"result_high_migration_%s.json",
			forecastUUID.String(),
		)

		lowMigrationForecastResults := helpers.ReadPrognosisResultFile(lowMigrationResultFileName)
		mediumMigrationForecastResults := helpers.ReadPrognosisResultFile(mediumMigrationResultFileName)
		highMigrationForecastResults := helpers.ReadPrognosisResultFile(highMigrationResultFileName)

		// Now build the response
		response := structs.DataResponse{
			LowMigrationData:    lowMigrationForecastResults,
			MediumMigrationData: mediumMigrationForecastResults,
			HighMigrationData:   highMigrationForecastResults,
		}

		// Now send back the response
		responseWriter.Header().Set("Content-Type", "text/json")
		encodingError := json.NewEncoder(responseWriter).Encode(response)
		if encodingError != nil {
			logger.WithError(encodingError).Error("Unable to encode the request error into json")
			return
		}
	}
}
