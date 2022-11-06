#!/usr/bin/env Rscript
library(argparser, quietly=TRUE)
library(prophet, quietly=TRUE)

# Create a new argparser

p <- arg_parser("Run a new forecast with a prophet-based model")

# Add cli arguments
p <- add_argument(p, "requestID", help="The request id prepended to all input and output files", type="character")

# Parse the cli parameters
argv <- parse_args(p)

# Build the required files
waterUsagesFile <- paste("water_usage_", argv$requestID, ".json", sep="")
currentPopulationFile <- paste("current_population_", argv$requestID, ".json", sep="")
lowPopulationMigrationFile <- paste("low_population_migration_", argv$requestID, ".json", sep="")
mediumPopulationMigrationFile <- paste("medium_population_migration_", argv$requestID, ".json", sep="")
highPopulationMigrationFile <- paste("high_population_migration_", argv$requestID, ".json", sep="")
lowMigrationResultFile <- paste("result_low_migration_", argv$requestID, ".json", sep="")
mediumMigrationResultFile <- paste("result_medium_migration_", argv$requestID, ".json", sep="")
highMigrationResultFile <- paste("result_high_migration_", argv$requestID, ".json", sep="")

# Read the file contents
realWaterUsages <- jsonlite::read_json(waterUsagesFile, simplifyVector = TRUE)
currentPopulation <- jsonlite::read_json(currentPopulationFile, simplifyVector = TRUE)
lowPopulationMigration <- jsonlite::read_json(lowPopulationMigrationFile, simplifyVector = TRUE)
mediumPopulationMigration <- jsonlite::read_json(mediumPopulationMigrationFile, simplifyVector = TRUE)
highPopulationMigration <- jsonlite::read_json(highPopulationMigrationFile, simplifyVector = TRUE)

# Start building the prophet model
model <- prophet()
model <- prophet::add_country_holidays(model, "DE")
modelWaterUsages <- fit.prophet(model, realWaterUsages)
futureDs <- prophet::make_future_dataframe(modelWaterUsages, 43, freq="year")
# Forecast the water usage values
forecastedWaterUsagesFull <- predict(modelWaterUsages, futureDs)
forecastedWaterUsagesFull <- forecastedWaterUsagesFull[-1, ]
forecastedWaterUsages <- forecastedWaterUsagesFull[c('ds', 'yhat', 'yhat_lower', 'yhat_upper')]
lowerBoundWaterUsages <- forecastedWaterUsagesFull[c('ds', 'yhat_lower')]
predictedWaterUsages <- forecastedWaterUsagesFull[c('ds', 'yhat')]
upperBoundWaterUsages <- forecastedWaterUsagesFull[c('ds', 'yhat_upper')]

# Merge the current and forecasted population values
lowPopulationMigration <- rbind(currentPopulation, lowPopulationMigration)
mediumPopulationMigration <- rbind(currentPopulation, mediumPopulationMigration)
highPopulationMigration <- rbind(currentPopulation, highPopulationMigration)



# Calculate the per-person usages for the low population migration data
lowerBoundLowMigrationUsages <- lowPopulationMigration
lowerBoundLowMigrationUsages$y <- NULL
lowerBoundLowMigrationUsages$values <- lowerBoundWaterUsages$yhat_lower / lowPopulationMigration$y

forecastedLowMigrationUsages <- lowPopulationMigration
forecastedLowMigrationUsages$y <- NULL
forecastedLowMigrationUsages$values <- predictedWaterUsages$yhat / lowPopulationMigration$y

upperBoundLowMigrationUsages <- lowPopulationMigration
upperBoundLowMigrationUsages$y <- NULL
upperBoundLowMigrationUsages$values <- upperBoundWaterUsages$yhat_upper / lowPopulationMigration$y

# Calculate the per-person usages for the medium population migration data
lowerBoundMediumMigrationUsages <- mediumPopulationMigration
lowerBoundMediumMigrationUsages$y <- NULL
lowerBoundMediumMigrationUsages$values <- lowerBoundWaterUsages$yhat_lower / mediumPopulationMigration$y

forecastedMediumMigrationUsages <- mediumPopulationMigration
forecastedMediumMigrationUsages$y <- NULL
forecastedMediumMigrationUsages$values <- predictedWaterUsages$yhat / mediumPopulationMigration$y

upperBoundMediumMigrationUsages <- mediumPopulationMigration
upperBoundMediumMigrationUsages$y <- NULL
upperBoundMediumMigrationUsages$values <- upperBoundWaterUsages$yhat_upper / mediumPopulationMigration$y

# Calculate the per-person usages for the high population migration data
lowerBoundHighMigrationUsages <- highPopulationMigration
lowerBoundHighMigrationUsages$y <- NULL
lowerBoundHighMigrationUsages$values <- lowerBoundWaterUsages$yhat_lower / highPopulationMigration$y

forecastedHighMigrationUsages <- highPopulationMigration
forecastedHighMigrationUsages$y <- NULL
forecastedHighMigrationUsages$values <- predictedWaterUsages$yhat / highPopulationMigration$y

upperBoundHighMigrationUsages <- highPopulationMigration
upperBoundHighMigrationUsages$y <- NULL
upperBoundHighMigrationUsages$values <- upperBoundWaterUsages$yhat_upper / highPopulationMigration$y

# Migrate the data back into single dataframes
lowMigrationPerPersonUsages <- lowerBoundLowMigrationUsages
lowMigrationPerPersonUsages$values <- NULL
lowMigrationPerPersonUsages$lower <- lowerBoundLowMigrationUsages$values
lowMigrationPerPersonUsages$forecast <- forecastedLowMigrationUsages$values
lowMigrationPerPersonUsages$upper <- upperBoundLowMigrationUsages$values

mediumMigrationPerPersonUsages <- lowerBoundMediumMigrationUsages
mediumMigrationPerPersonUsages$values <- NULL
mediumMigrationPerPersonUsages$lower <- lowerBoundMediumMigrationUsages$values
mediumMigrationPerPersonUsages$forecast <- forecastedMediumMigrationUsages$values
mediumMigrationPerPersonUsages$upper <- upperBoundMediumMigrationUsages$values

highMigrationPerPersonUsages <- lowerBoundHighMigrationUsages
highMigrationPerPersonUsages$values <- NULL
highMigrationPerPersonUsages$lower <- lowerBoundHighMigrationUsages$values
highMigrationPerPersonUsages$forecast <- forecastedHighMigrationUsages$values
highMigrationPerPersonUsages$upper <- upperBoundHighMigrationUsages$values 

jsonlite::write_json(lowMigrationPerPersonUsages, lowMigrationResultFile, digits = 10)
jsonlite::write_json(mediumMigrationPerPersonUsages, mediumMigrationResultFile, digits = 10)
jsonlite::write_json(highMigrationPerPersonUsages, highMigrationResultFile, digits = 10)
