package mariadb

import (
	"golang-project-template/entity"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entity.YourEntity{},
	)

	if err != nil {
		return err
	}

	return nil
}
