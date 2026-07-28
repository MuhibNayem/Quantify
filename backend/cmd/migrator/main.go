package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dir := flag.String("dir", "migrations", "migrations directory")
	cmd := flag.String("cmd", "up", "up | down | version")
	dsn := flag.String("dsn", os.Getenv("DATABASE_URL"), "Postgres DSN")
	flag.Parse()

	m, err := migrate.New(fmt.Sprintf("file://%s", *dir), *dsn)
	if err != nil {
		log.Fatal(err)
	}
	switch *cmd {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "version":
		v, dirty, _ := m.Version()
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
		return
	}
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
	fmt.Println("OK")
}
