package database

import (
	"log"
	"os"

	"envie-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable not set")
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            false,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connection established")

	log.Println("Running migrations...")
	if err := db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.ConfigItem{},
		&models.SecretManagerConfig{},
		&models.UserIdentity{},

		&models.Organization{},
		&models.OrganizationUser{},
		&models.Team{},
		&models.TeamUser{},
		&models.TeamProject{},

		&models.PendingKeyRotation{},
		&models.KeyRotationApproval{},

		&models.ProjectFile{},

		&models.LinkingCode{},

		&models.ProjectToken{},
		// RefreshToken table no longer needed - using stateless JWTs
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	DB = db
}
