package mariadb

import (
	"golang-project-template/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Connection *gorm.DB

func ConnectDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(config.LoadDataSourceName()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	Connection = db

	return Connection, nil
}
