package connection

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func createLogger() *zap.Logger {
	config := zap.Config{
		Level:    zap.NewAtomicLevelAt(zap.InfoLevel),
		Encoding: "json",
		OutputPaths: []string{
			"logs/logs.txt",
		},
		EncoderConfig: zap.NewProductionEncoderConfig(),
	}

	return zap.Must(config.Build())
}

func UpdateQuery(query string, amount int) {
	logger := createLogger()
	defer logger.Sync()
	logger.Info("Application started")

	db, err := sqlx.Connect("postgres", "user=postgres dbname=demo sslmode=disable password=secret host=localhost")
	logger.Info("Trying to connect to database")
	if err != nil {
		log.Fatalln(err)
		logger.Info("Error occurred while trying to connect to the database")
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
		logger.Info("database not reachable")
	} else {
		fmt.Println("[+] Successfully connected to database")
		logger.Info("Successfully connected to database")
	}

	fmt.Println("[*] running query..")
	logger.Info("Running query")

	_, err = db.Exec(query, amount)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("[+] Data successfully updated")
	logger.Info("Successfully executed query")
	logger.Info("Application closed")
}
