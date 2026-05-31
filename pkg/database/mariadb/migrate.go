package mariadb

import (
	"github.com/azmiagr/golang-project-template/entity"

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
