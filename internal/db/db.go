package db

import (
    "log"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dsn string) error {
    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return err
    }
	if err := DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return err
	}

    err = DB.AutoMigrate(
        &User{},
        &Transaction{},
        &Product{},
        &Purchase{},
        &LotParticipant{},
        &Banner{},
        &AdminLog{},
        &MediaUpload{},
        &Notification{},
        &BroadcastMessage{},
        &BroadcastHistory{},
<<<<<<< HEAD
        &Setting{},
=======
>>>>>>> 7259f56d86e51521de5028adf64898f696632916
    )
    if err != nil {
        return err
    }

    log.Println("Database migrated successfully")
    return nil
}
