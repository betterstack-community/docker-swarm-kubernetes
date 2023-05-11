package config

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func init() {
	var err error
	DB, err = sql.Open("postgres", "postgres://postgres:example@localhost/blog?sslmode=disable")
	if err != nil {
		panic(err)
	}

	if err = DB.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("You connected to your database.")

	//signBytes, err := ioutil.ReadFile(PrivKeyPath)
	//if err != nil {
	//	panic(err)
	//}
	//SignKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	//if err != nil {
	//	panic(err)
	//}
	//verifyBytes, err := ioutil.ReadFile(PubKeyPath)
	//if err != nil {
	//	panic(err)
	//}
	//VerifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	//if err != nil {
	//	panic(err)
	//}
}
