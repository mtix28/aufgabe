package connection

import (
	"fmt"
	"log"

	"github.com/spf13/viper"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// config for zap logger
func createLogger() *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	config := zap.Config{
		Level:    zap.NewAtomicLevelAt(zap.InfoLevel),
		Encoding: "json",
		OutputPaths: []string{
			"logs/logs.txt",
		},
		EncoderConfig: encoderConfig,
	}

	return zap.Must(config.Build())
}

func UpdateQuery(query string, amount int) {
	// create logger
	logger := createLogger()
	defer logger.Sync()
	logger.Info("Application started")

	//load .env
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	logger.Info("Loaded .env file")

	user := viper.Get("DB_USER")
	dbname := viper.Get("DB_NAME")
	password := viper.Get("DB_PASSWORD")
	host := viper.Get("DB_HOST")
	connection_string := fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s host=%s", user, dbname, password, host)
	logger.Info("Created connection string")

	db, err := sqlx.Connect("postgres", connection_string)
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
