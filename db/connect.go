package connection

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func UpdateQuery(query string, amount int) {
	db, err := sqlx.Connect("postgres", "user=postgres dbname=demo sslmode=disable password=secret host=localhost")
	if err != nil {
		log.Fatalln(err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("[+] Successfully connected to database")
	}

	log.Println("[*] running query..")

	_, err = db.Exec(query, amount)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("[+] Data successfully updated")
}
