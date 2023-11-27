package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error

	viper.AddConfigPath("../..")
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	err = viper.ReadInConfig()

	if err != nil {
		log.Fatal("Error loading .env file ", err)
	}

	// Get database driver and source from environment variables
	dbDriver := viper.Get("DB_DRIVER").(string)
	dbSource := viper.Get("DB_SOURCE").(string)

	testDB, err = sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db ", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
