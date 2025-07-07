package database

import (
	"context"
	"database/sql"
	"time"

	// go-pg db driver
	_ "github.com/lib/pq"
)

var (
	approvalPGConnection *sql.DB
	dev                  *sql.DB
)

// GetDatabaseConnection to get the connetion
func GetDatabaseConnection(database string) (*sql.DB, error) {

	//start := time.Now()
	// defer func() {
	//      fmt.Print("Ping :- ", time.Since(start))
	// }()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var err error

	switch database {

	case "approvalPG":
		if approvalPGConnection != nil {
			err = approvalPGConnection.PingContext(ctx)
		} else {
			err = GetPostGresConnection(database)
		}
		return approvalPGConnection, err
	case "dev":
		if dev != nil {
			err = dev.PingContext(ctx)
		} else {
			err = GetPostGresConnection(database)
		}
		return dev, err
	}

	return nil, err
}

// GetPostGresConnection to get the postgress db connection
func GetPostGresConnection(database string) error {

	var err error
	username := "bi"
	password := "bipass4impaypg"
	host := "34.93.67.72" //MasterNode of ApporvalPG
	port := "5432"
	dbName := "approvalpg"
	sslMode := "disable"

	if database == "dev" {
		//host = "34.100.170.199"   // ReadNode of LIVE ApprovalPG
		// host = "34.93.201.174"  old one
		host = "34.100.240.197"
		port = "5432"
		dbName = "mesh_glusr"
	}

	pgConnString := "host=" + host + " port=" + port + " user=" + username + " dbname=" + dbName + " sslmode=" + sslMode + " password=" + password
	connection, err := sql.Open("postgres", pgConnString)
	connection.SetConnMaxLifetime(4 * time.Hour)
	connection.SetMaxIdleConns(20)
	connection.SetMaxOpenConns(200)

	if database == "dev" {
		dev = connection
	} else {
		approvalPGConnection = connection
	}

	return err
}
