package enums

// MigrationLevel is a type constraint
type MigrationLevel string

const (
	LowMigrationLevel    MigrationLevel = "low"
	MediumMigrationLevel MigrationLevel = "medium"
	HighMigrationLevel   MigrationLevel = "high"
)
