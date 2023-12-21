package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

type Wallet struct {
	Address       string `json:"address"`
	Blockchain    string `json:"blockchain"`
	Type          string `json:"type"`
	NetWorthUSD   string `json:"net_worth_usd"`
	FirstIn       string `json:"first_in"`
	FirstOut      string `json:"first_out"`
	LastIn        string `json:"last_in"`
	LastOut       string `json:"last_out"`
	WalletAgeDays int    `json:"wallet_age_days"`
}

type Data struct {
	Operations []Transactions `json:"operations"`
}

type Transactions struct {
	Datetime        int32     `json:"timestamp"`
	Sender          string    `json:"from"`
	Receiver        string    `json:"to"`
	TransactionHash string    `json:"transactionHash"`
	Value           string    `json:"value"`
	TokenInfo       TokenInfo `json:"tokenInfo"`
}

type TokenInfo struct {
	Symbol   string `json:"symbol"`
	Decimals string `json:"decimals"`
}

const create string = `
	CREATE TABLE IF NOT EXISTS accounts(
		"id" INTEGER PRIMARY KEY AUTOINCREMENT,
		"address"       	TEXT NOT NULL,        
		"blockchain"    	TEXT NOT NULL,            
		"type"          	TEXT NOT NULL,       
		"net_worth_usd"	TEXT NOT NULL,   	     
		"first_in"       	DATETIME NOT NULL,        
		"first_out"      	DATETIME NOT NULL,        
		"last_out"        DATETIME NOT NULL,     
		"last_in"      	TEXT,    
		"wallet_age_days" INTEGER         
	);


	CREATE TABLE IF NOT EXISTS tx (
		"acc_id" INTEGER,
		"time" DATETIME NOT NULL,
		"tx_hash" TEXT NOT NULL,
		"value" REAL NOT NULL,
		"symbol" TEXT NOT NULL,
		"sender" TEXT NOT NULL,
		"receiver" TEXT NOT NULL,
		FOREIGN KEY ("acc_id") REFERENCES accounts ("id")
	);`

func NewDatabase() (*Database, error) {
	db, err := sql.Open("sqlite3", "./storage/tx.db")
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(create); err != nil {
		return nil, err
	}
	return &Database{db: db}, nil
}

func (d *Database) GetWalletTx(address string) {
	var wallet Wallet
	resp, err := http.Get("https://leap.oraclus.com/v1/address/ethereum/" + address)
	if err != nil {
		log.Println(err)
		return
	}
	err = json.NewDecoder(resp.Body).Decode(&wallet)
	if err != nil {
		log.Println(err)
		return
	}
	res, err := d.db.Exec("INSERT INTO accounts VALUES(NULL,?,?,?,?,?,?,?,?,?);", wallet.Address, wallet.Blockchain, wallet.Type, wallet.NetWorthUSD, wallet.FirstIn, wallet.FirstOut, wallet.LastIn, wallet.LastOut, wallet.WalletAgeDays)
	if err != nil {
		log.Println("faield to add wallet to database: ", err)
		return
	}
	var id int64
	if id, err = res.LastInsertId(); err != nil {
		log.Println(err)
	}

	var data Data
	response, err := http.Get("https://api.ethplorer.io/getAddressHistory/" + address + "?apiKey=EK-fiMNj-WdkpNfm-dY9Af&limit=1000")
	if err != nil {
		log.Println(err)
		return
	}

	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		log.Println(err)
		return
	}

	for i := 0; i < len(data.Operations); i++ {
		_, err := d.db.Exec("INSERT INTO tx VALUES(?,?,?,?,?,?,?);", id, data.Operations[i].Datetime, data.Operations[i].TransactionHash, data.Operations[i].Value, data.Operations[i].TokenInfo.Symbol, data.Operations[i].Sender, data.Operations[i].Receiver)
		if err != nil {
			log.Println("failed to add row: ", err)
		}
	}

}

type FrontData struct {
	Datetime string
	TxHash   string
	Amount   string
	Receiver string
	Symbol   string
}

func (d *Database) GetAllTransactions(address string) ([]FrontData, error) {
	row := d.db.QueryRow("SELECT id FROM accounts WHERE address='" + address + "'")

	var id int
	if err := row.Scan(&id); err != nil{
		if !errors.Is(err, sql.ErrNoRows){
			return nil, err
		}
		d.GetWalletTx(address)
		
	}
	fmt.Println(id)
	
	rows, err := d.db.Query("SELECT time, tx_hash, value, receiver, symbol FROM tx WHERE acc_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := []FrontData{}

	for rows.Next() {
		i := FrontData{}
		if err := rows.Scan(&i.Datetime, &i.TxHash, &i.Amount, &i.Receiver, &i.Symbol); err != nil {
			return nil, err
		}
		data = append(data, i)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return data, nil
}
