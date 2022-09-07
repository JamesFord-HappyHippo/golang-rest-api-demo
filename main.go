/*
REST API DEMO

Methods :
createNewCompany, returnAllCompany, returnSingleCompany, updateCompany, homepage, deleteCompany, handleRequests, connectToDB, main
*/
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type (

	// App controlls the rest API demo app
	App struct {
		DBType   string
		Router   *mux.Router
		Database *sql.DB
		logger   *log.Logger
	}

	// Company contains the data to be details for data to be stored into DB
	// prepare Company data

	Company struct {
		Client_ID          int    `json:"Client_ID"`
		Company_ID         int    `json:"Company_ID"`
		Company_Name       string `json:"Company_Name"`
		ASIC               string `json:"ASIC"`
		Flight_Risk_Status string `json:"Flight_Risk_Status"`
		Recruit_Status     string `json:"Recruit_Status"`
		Total_Flight_Risk  string `json:"Total_Flight_Risk"`
		Total_Backfill     string `json:"Total_Backfill"`
		Create_Date        string `json:"Create_Date"`
		Last_Update        string `json:"Last_Update"`
		Data_As_Of_Date    string `json:"Data_As_Of_Date"`
	}
)

//	POST /createNewCompany
//	payload : Company struct
//
// creates new Company entry to DB
func (app *App) createNewCompany(w http.ResponseWriter, r *http.Request) {

	var (
		query   string
		Company Company
	)

	app.logger.Println("Endpoint hit : createNewCompany")
	// get the payload from request
	err := json.NewDecoder(r.Body).Decode(&Company)
	if err != nil {
		app.logger.Println(err)
	}

	if app.DBType == "mysql" {
		query = "INSERT INTO Company_Detail (Client_ID, Company_ID, Company_Name, ASIC, Flight_Risk_Status, Recruit_Status, Total_Flight_Risk, Total_Backfill, Create_Date, Last_Update, Data_as_of_Date) VALUES (?,?,?,?,?,?,?,?,?,?,?)"
	} else if app.DBType == "postgres" {
		query = "INSERT INTO Company_Detail (title, descr, content) VALUES ($1,$2,$3)"
	}

	// insert data into DB
	response, err := app.Database.Exec(query, Company.Client_ID, Company.Company_ID, Company.Company_Name)
	// if there is an error inserting, handle it
	if err != nil {
		app.logger.Println(err.Error())
		return
	}
	app.logger.Print(response.RowsAffected())
	app.logger.Println("inserted new record to DB")

	// return the added Company
	json.NewEncoder(w).Encode(Company)
}

//	GET /returnAllCompany_Detail
//	query params : id (last displayed ID for pagination), limit (max entry count in display)
//	response     : Company struct array
//
// get all the Company_Detail from DB
func (app *App) returnAllCompany_Detail(w http.ResponseWriter, r *http.Request) {

	var (
		query          string
		queryParams    []interface{}
		Company_Detail []Company
	)

	app.logger.Println("Endpoint hit : returnAllCompany_Detail")

	// get the id and limit from param
	lastID := r.URL.Query().Get("id")
	limit := r.URL.Query().Get("limit")

	// if last id is empty, set as 0
	if lastID == "" {
		lastID = "0"
	}

	// if limti is empty, get all entries else get all entries with limit
	if limit == "" {

		if app.DBType == "mysql" {
			query = "SELECT * FROM Company_Detail WHERE Company_ID > ? ORDER BY Company_ID ASC"
		} else if app.DBType == "postgres" {
			query = "SELECT * FROM Company_Detail WHERE id > $1 ORDER BY id ASC"
		}
		queryParams = append(queryParams, lastID)
	} else {

		if app.DBType == "mysql" {
			query = "SELECT * FROM Company_Detail WHERE Company_ID > ? ORDER BY Company_ID ASC LIMIT ?"
		} else if app.DBType == "postgres" {
			query = "SELECT * FROM Company_Detail WHERE id > $1 ORDER BY id ASC LIMIT $2"
		}
		queryParams = append(queryParams, lastID, limit)
	}
	app.logger.Println(query, queryParams)

	// insert data into DB
	response, err := app.Database.Query(
		query,
		queryParams...,
	)
	// if there is an error inserting, handle it
	if err != nil {
		app.logger.Println(err.Error())
		return
	}
	defer response.Close()

	// get all records until all are read
	for response.Next() {

		var Company Company

		// get data from DB for Company fields
		err = response.Scan(
			&Company.Client_ID,
			&Company.Company_ID,
			&Company.Company_Name,
			&Company.ASIC,
		)
		// if there is an error inserting, handle it
		if err != nil {
			app.logger.Println(err.Error())
			return
		}

		// append to final list of Company_Detail
		Company_Detail = append(Company_Detail, Company)
	}
	app.logger.Printf("Company : %+v\n", Company_Detail)

	// generate JSON resopnse
	err = json.NewEncoder(w).Encode(Company_Detail)
	if err != nil {
		app.logger.Println(err)
	}
	app.logger.Println("Endpoint hit : return all Company_Detail")
}

//	GET /returnSingleCompany/{id}
//	url params : id (Company ID to be retrieved)
//	response   : Company struct
//
// return a selected Company value from DB
func (app *App) returnSingleCompany(w http.ResponseWriter, r *http.Request) {

	var (
		query   string
		Company Company
	)

	app.logger.Println("Endpoint hit : returnSingleCompany")
	// get url path parameters
	vars := mux.Vars(r)
	key := vars["id"]

	if app.DBType == "mysql" {
		query = "SELECT * FROM Company_Detail WHERE Company_ID=?"
	} else if app.DBType == "postgres" {
		query = "SELECT * FROM Company_Detail WHERE id=$1"
	}

	// insert data into DB
	response, err := app.Database.Query(query, key)
	// if there is an error inserting, handle it
	if err != nil {
		app.logger.Println(err.Error())
		return
	}
	defer response.Close()

	// iterate until entries from db are read
	for response.Next() {

		// scan and get Company fields value
		err = response.Scan(
			&Company.Client_ID,
			&Company.Company_ID,
			&Company.Company_Name,
			&Company.ASIC,
		)
		// if there is an error inserting, handle it
		if err != nil {
			app.logger.Println(err.Error())
			return
		}
	}
	app.logger.Printf("Company : %+v\n", Company)

	// if Company ID is not empty, return JSON response
	if Company.Company_ID != 0 {
		json.NewEncoder(w).Encode(Company)
	} else {
		http.Error(w, "no record", http.StatusNotFound)
	}
}

//	PUT /updateCompany/{id}
//	url params : id (Company ID to be retrieved)
//
// update the Company for a given Company ID
func (app *App) updateCompany(w http.ResponseWriter, r *http.Request) {

	var (
		query          string
		updatedCompany Company
	)

	app.logger.Println("Endpoint hit : updateCompany")
	// get the path parameter
	vars := mux.Vars(r)
	key := vars["Company_ID"]

	// get the payload data for Company
	err := json.NewDecoder(r.Body).Decode(&updatedCompany)
	if err != nil {
		app.logger.Println(err)
	}

	if app.DBType == "mysql" {
		query = "UPDATE Company_Detail SET Client_ID=? , Company_ID=?, Company_Name=?, ASIC=?, Flight_Risk_Status=?, Recruit_Status=?, Total_Flight_Risk=?, Total_Backfill=?, Create_Date=?, Last_Update=?, Data_as_of_Date=? WHERE id=?"
	} else if app.DBType == "postgres" {
		query = "UPDATE Company_Detail SET title=$1, descr=$2, content=$3 WHERE id=$4"
	}

	// update data in DB
	response, err := app.Database.Exec(query, updatedCompany.Client_ID, updatedCompany.Company_ID, updatedCompany.Company_Name, key)
	// if there is an error inserting, handle it
	if err != nil {
		app.logger.Println(err.Error())
		return
	}
	app.logger.Print(response.RowsAffected())
	app.logger.Println(" DB update performed.")

	// return the JSON response for added Company
	json.NewEncoder(w).Encode(updatedCompany)
}

//	DELETE /deleteCompany/{id}
//	url params : id (Company ID to be retrieved)
//
// remove an Company from DB
func (app *App) deleteCompany(w http.ResponseWriter, r *http.Request) {

	var query string

	app.logger.Println("Endpoint hit : deleteCompany")
	// get url path parameter
	vars := mux.Vars(r)
	key := vars["Company_ID"]

	if app.DBType == "mysql" {
		query = "DELETE FROM Company_Detail WHERE Company_ID=?"
	} else if app.DBType == "postgres" {
		query = "DELETE FROM Company_Detail WHERE id=$1"
	}

	// insert data into DB
	response, err := app.Database.Exec(query, key)
	// if there is an error inserting, handle it
	if err != nil {
		app.logger.Println(err.Error())
		return
	}
	app.logger.Print(response.RowsAffected())
	app.logger.Println(" DB delete performed.")
}

//	ANY /homepage
//
// home page of web server
func (app *App) homepage(w http.ResponseWriter, r *http.Request) {

	app.logger.Println("Endpoint hit : homepage")
	fmt.Fprint(w, `
- POST /Company
  - Add new Company to DB
  - payload :
    {
		Client_ID 	(string)
		Company_ID 	(string)
		Company_Name (string)
		ASIC 		(string)
		Flight_Risk_Status (string)
		Recruit_Status (string)
		Total_Flight_Risk (string)
		Total_Backfill   (string)
		Create_Date (string)
		Last_Update   (string)
		Data_As_Of_Date  (string)
    }

- PUT /Company/{id}
  - Update an existing Company DB
  - query param : id (Company id from GET API)
  - payload :
  {
	  Client_ID 	(string)
	  Company_ID 	(string)
	  Company_Name (string)
	  ASIC 		(string)
	  Flight_Risk_Status (string)
	  Recruit_Status (string)
	  Total_Flight_Risk (string)
	  Total_Backfill   (string)
	  Create_Date (string)
	  Last_Update   (string)
	  Data_As_Of_Date  (string)
  }

- DELETE /Company/{id}
  - Deletes an entry from DB
  - query param : id (Company id from GET API)

- GET /Company/{id}
  - Retrieves Company data from DB for a given ID
  - query param : id (Company id from GET API) 

- GET /Company_Detail
  - retrives all Company_Detail from DB
  - query params : id (last ID from previous GET call for pagination), limit (max entry per page)
  - response : list of Company_Detail
`)
}

// http handler methods init
func handleRequests(app *App, port string) {

	// start the gorilla mux router
	app.Router = mux.NewRouter().StrictSlash(true)

	// http routes
	app.Router.HandleFunc("/", app.homepage)
	app.Router.HandleFunc("/Company_Detail", app.returnAllCompany_Detail).Methods("GET")
	app.Router.HandleFunc("/Company_Detail", app.createNewCompany).Methods("POST")
	app.Router.HandleFunc("/Company_Detail/{Company_ID}", app.updateCompany).Methods("PUT")
	app.Router.HandleFunc("/Company_Detail/{Company_ID}", app.deleteCompany).Methods("DELETE")
	app.Router.HandleFunc("/Company_Detail/{Company_ID}", app.returnSingleCompany).Methods("GET")

	// start the server on port
	app.logger.Fatal(http.ListenAndServe(":"+port, app.Router))
}

// establish DB connection for mysql DB
func connectToDB(dbType, connectionString string, logger *log.Logger) (db *sql.DB, err error) {

	// establish new db connection
	db, err = sql.Open(dbType, connectionString)

	// if there is an error opening the connection, handle it
	if err != nil {
		logger.Println(err.Error())
		return
	}

	// execute a ping on DB
	err = db.Ping()

	// if there is an error opening the connection, handle it
	if err != nil {
		logger.Println(err.Error())
		return
	}
	logger.Println("Established "+dbType+" DB connection for ", connectionString)
	return
}

// main function
func main() {

	var connectionString string

	//dbType := flag.String("mysql")
	//dbUser := flag.String("admin")
	//dbPass := flag.String("44_FUNtime")
	//dbHost := flag.String("happy1.cwkfm0ctmqb3.us-east-2.rds.amazonaws.com")
	//dbPort := flag.String("3306")
	//	dbName := flag.String("Happy1")
	//	port := flag.String("7777")
	//	flag.Parse()

	// based on the db type set the connection string
	const dbType = "mysql"

	//	connectionString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", *dbUser, *dbPass, *dbHost, *dbPort, *dbName)
	connectionString = "admin:44_FUNtime@tcp(happy1.cwkfm0ctmqb3.us-east-2.rds.amazonaws.com:3306)/Happy1"

	//} else if *dbType == "postgres" {

	//	connectionString = fmt.Sprintf(
	//		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
	//		*dbHost, *dbPort, *dbUser, *dbPass, *dbName,
	//	)
	//}

	// store the log file data to log file
	logFile, _ := os.OpenFile(
		"./restful_api.log",
		os.O_TRUNC|os.O_CREATE|os.O_RDWR,
		os.ModePerm,
	)

	logger := log.New(
		logFile,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile,
	)

	// connect to DB
	dbConn, err := connectToDB("mysql", connectionString, logger)
	if err != nil {
		log.Println(err)
	}

	// set new router
	app := &App{
		DBType:   "mysql",
		Router:   mux.NewRouter().StrictSlash(true),
		Database: dbConn,
		logger:   logger,
	}

	// defer the close till after the main function has finished
	// executing
	defer app.Database.Close()

	// initialize the routes for rest API server
	handleRequests(app, "7777")
}
