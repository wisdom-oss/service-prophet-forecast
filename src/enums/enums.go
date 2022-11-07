package enums

type MigrationLevel string

// Enumerations for specifying the migration levels available for the prognosis
const (
	LowMigrationLevel    MigrationLevel = "low"
	MediumMigrationLevel MigrationLevel = "medium"
	HighMigrationLevel   MigrationLevel = "high"
)
