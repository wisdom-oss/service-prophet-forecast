package utils

import (
	"database/sql"
	"fmt"
	"microservice/structs"
)

// ReadDataForProphet accepts a Rows object which has the following columns: year,
// data. This data is then transformed into a dataset to handle the access better
func ReadDataForProphet(rows *sql.Rows) ([]structs.InputDataPoint, error) {
	var dataset []structs.InputDataPoint
	for rows.Next() {
		var year int
		var population float64

		scanError := rows.Scan(&year, &population)
		if scanError != nil {
			logger.WithError(scanError).Error("An error occurred while parsing the data from the database")
			return nil, scanError
		}

		dataset = append(
			dataset, structs.InputDataPoint{
				Date:  fmt.Sprintf(`%d-12-31`, year),
				Value: population,
			},
		)
	}
	return dataset, nil
}
