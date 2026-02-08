package postgresql

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDb struct {
}

var db *gorm.DB

func NewDb(connectionString string) *PostgresDb {
	conn, err := gorm.Open(postgres.Open(connectionString))
	if err != nil {
		fmt.Print(err.Error())
		panic("Couldn't open database connection")
	}

	db = conn
	//automigrate
	return &PostgresDb{}
}
