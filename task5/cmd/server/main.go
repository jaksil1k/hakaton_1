package main

import (
	"log"

	"github.com/tunx321/task5/internal/db"
	"github.com/tunx321/task5/internal/service"
	"github.com/tunx321/task5/internal/transport"
)

func main() {
	db, err := db.NewDatabase()
	if err != nil {
		log.Fatal(err)
	}
	
	//db.GetWalletTx("0xfd7b50da848f3e3263cfE7f5A2De23530C2Cd9FB")
	// db.GetAllTransactions("0xfd7b50da848f3e3263cfE7f5A2De23530C2Cd9FB")
	txSrv := service.NewService(db)
	hanlder := transport.NewHandler(txSrv)

	if err := hanlder.Serve(); err != nil {
		log.Fatal(err)
	}
}
