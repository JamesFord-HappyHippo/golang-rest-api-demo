package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
)

//const PostgresConn = "host=localhost port=5432 user=postgres password=mysecretpassword dbname=postgres sslmode=disable"
const MySQLConn = "admin:44_FUNtime@tcp(happy1.cwkfm0ctmqb3.us-east-2.rds.amazonaws.com:3306)/Happy1"

func TestConnectToDB(t *testing.T) {

	logFile, _ := os.OpenFile(
		"./restful_api.log",
		os.O_TRUNC|os.O_CREATE|os.O_RDWR,
		os.ModePerm,
	)

	logger := log.New(logFile, "INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	// positive case for mysql DB conn
	dbConn, err := connectToDB("mysql", MySQLConn, logger)
	if err != nil || dbConn.Ping() != nil {
		t.Errorf("Failed to establish mysql DB connection")
	}
	dbConn.Close()

	// negative case for mysql DB conn
	dbConn, err = connectToDB("mysql", "admin:44_FUNtime@tcp(happy1.cwkfm0ctmqb3.us-east-2.rds.amazonaws.com:3306)/Happy1", logger)
	if err == nil || dbConn.Ping() == nil {
		t.Errorf("DB Connection established for wrong user name / password")
		dbConn.Close()
	}

	// negative case for mysql DB conn
	dbConn, err = connectToDB("mysql", "admin:44_FUNtime@tcp(happy1.cwkfm0ctmqb3.us-east-2.rds.amazonaws.com:3306)/Happy1", logger)
	if err == nil || dbConn.Ping() == nil {
		t.Errorf("DB Connection established for db host/port")
	}
	dbConn.Close()

	// positive case for postgres DB conn
	dbConn, err = connectToDB("postgres", PostgresConn, logger)
	if err != nil || dbConn.Ping() != nil {
		t.Errorf("Failed to establish mysql DB connection")
	}
	dbConn.Close()

	// negative case for postgres DB conn
	dbConn, err = connectToDB("postgres", "host=localhost port=5432 user=unknown password=unknown dbname=postgres sslmode=disable", logger)
	if err == nil || dbConn.Ping() == nil {
		t.Errorf("DB Connection established for wrong user name / password")
	}
	dbConn.Close()

	// negative case for postgres DB conn
	dbConn, err = connectToDB("postgres", "host=localhost port=1234 user=postgres password=mysecretpassword dbname=postgres sslmode=disable", logger)
	if err == nil || dbConn.Ping() == nil {
		t.Errorf("DB Connection established for db host/port")
	}
	dbConn.Close()
}

func initTestModule(db, connectionString string) (app *App) {

	logFile, _ := os.OpenFile(
		"./restful_api.log",
		os.O_TRUNC|os.O_CREATE|os.O_RDWR,
		os.ModePerm,
	)

	logger := log.New(logFile, "INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	// connect to DB
	dbConn, err := connectToDB(db, connectionString, logger)
	if err != nil {
		log.Println(err)
	}

	// set new router
	app = &App{
		DBType:   db,
		Router:   mux.NewRouter().StrictSlash(true),
		Database: dbConn,
		logger:   logger,
	}

	return
}

func TestHomepage(t *testing.T) {

	const homepageResponse = `
- POST /company
  - Add new article to DB
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

- PUT /company/{id}
  - Update an existing article DB
  - query param : id (article id from GET API)
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

- DELETE /company/{id}
  - Deletes an entry from DB
  - query param : id (article id from GET API)

- GET /company/{id}
  - Retrieves article data from DB for a given ID
  - query param : id (article id from GET API) 

- GET /company
  - retrives all articles from DB
  - query params : id (last ID from previous GET call for pagination), limit (max entry per page)
  - response : list of articles
`

	// creating new request to homepage
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// start the DB connection and router for app
	app := initTestModule("mysql", MySQLConn)

	// if db connection fails, add as fatal error
	if app.Database == nil || app.Database.Ping() != nil {
		t.Fatal(err)
	}

	// new recorder for capturing response from request
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.homepage)

	// serve http call on request
	handler.ServeHTTP(rr, req)

	// checking http status code if 200
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response string matches expected response
	if rr.Body.String() != homepageResponse {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), homepageResponse)
	}
}

func TestCreateNewCompany(t *testing.T) {

	for _, dbType := range []string{"mysql", "postgres"} {

		var (
			connectionString string
			responseArticle  company
		)

		if dbType == "mysql" {
			connectionString = MySQLConn
		} else {
			connectionString = PostgresConn
		}

		// start the DB connection and router for app
		app := initTestModule(
			dbType,
			connectionString,
		)

		// prepare article data
		company := Company{
			Client_ID 	"2399029309299"
			Company_ID 	"21271d86-baf1-40c3-b803-77b526f8e4c5"
			Company_Name "TEST_GO"
			ASIC 		"1234"
			Flight_Risk_Status "Test"
			Recruit_Status "Test"
			Total_Flight_Risk "1234"
			Total_Backfill   "12345"
			Create_Date "2021-02-25"
			Last_Update   "2021-02-25"
			Data_As_Of_Date  "2021-02-25"
		}

		// convert the article data as json
		payload, err := json.Marshal(company)
		if err != nil {
			t.Fatal(err)
		}

		// creating new request to homepage
		req, err := http.NewRequest("POST", "/company", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}

		// if db connection fails, add as fatal error
		if app.Database == nil || app.Database.Ping() != nil {
			t.Fatal(err)
		}

		// new recorder for capturing response from request
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.createNewArticle)

		// serve http call on request
		handler.ServeHTTP(rr, req)

		// checking http status code if 200
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		// decode the response body to a new article struct
		json.NewDecoder(rr.Body).Decode(&responseArticle)

		// Check the response string matches expected response
		if responseCompany.Id != 0 && responseCompany.Company_Name != company.Company_Name {
			t.Errorf("handler returned unexpected body: got %+v want %+v",
				responseCompany, company)
		}
	}
}

func TestReturnAllArticles(t *testing.T) {

	for _, dbType := range []string{"mysql", "postgres"} {

		var (
			connectionString string
			responseCompany  company
		)

		if dbType == "mysql" {
			connectionString = MySQLConn
		} else {
			connectionString = PostgresConn
		}

		// start the DB connection and router for app
		app := initTestModule(
			dbType,
			connectionString,
		)

		// prepare article data
		company := Company{
			Client_ID 	"2399029309299"
			Company_ID 	"21271d86-baf1-40c3-b803-77b526f8e4c5"
			Company_Name "TEST_GO"
			ASIC 		"1234"
			Flight_Risk_Status "Test"
			Recruit_Status "Test"
			Total_Flight_Risk "1234"
			Total_Backfill   "12345"
			Create_Date "2021-02-25"
			Last_Update   "2021-02-25"
			Data_As_Of_Date  "2021-02-25"
		}

		// convert the article data as json
		payload, err := json.Marshal(article)
		if err != nil {
			t.Fatal(err)
		}

		// creating new request to homepage
		req, err := http.NewRequest("GET", "/company", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}

		// if db connection fails, add as fatal error
		if app.Database == nil || app.Database.Ping() != nil {
			t.Fatal(err)
		}

		// new recorder for capturing response from request
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.createNewCompany)

		// serve http call on request
		handler.ServeHTTP(rr, req)

		// checking http status code if 200
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		// decode the response body to a new article struct
		json.NewDecoder(rr.Body).Decode(&responseCompany)

		// Check the response string matches expected response
		if responseCompany.Id != 0 && responseCompany.Company_Name != company.Company_Name {
			t.Errorf("handler returned unexpected body: got %+v want %+v",
				responseCompany, company)
		}
	}
}
