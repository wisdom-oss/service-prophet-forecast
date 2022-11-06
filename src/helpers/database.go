package helpers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"microservice/structs"
)

// ReadDataForProphet accepts a Rows object which has the following columns: year,
// data. This data is then transformed into a dataset to handle the access better
func ReadDataForProphet(rows *sql.Rows) ([]structs.ProphetDataset, error) {
	var dataset []structs.ProphetDataset
	for rows.Next() {
		var year int
		var population float64

		scanError := rows.Scan(&year, &population)
		if scanError != nil {
			logger.WithError(scanError).Error("An error occurred while parsing the data from the database")
			return nil, scanError
		}

		dataset = append(
			dataset, structs.ProphetDataset{
				Date:  fmt.Sprintf(`%d-12-31`, year),
				Value: population,
			},
		)
	}
	return dataset, nil
}

func WriteDataToFile(content any, filename string) (int, error) {
	file, fileCreationError := os.Create(filename)
	if fileCreationError != nil {
		return -1, fileCreationError
	}

	fileContents, jsonMarshalError := json.Marshal(content)

	if jsonMarshalError != nil {
		return -1, jsonMarshalError
	}

	bytesWritten, writeError := file.Write(fileContents)

	if writeError != nil {
		return -1, writeError
	}

	return bytesWritten, nil

}
