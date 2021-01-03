package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// Route struct
type Route struct {
	DriverID string      `json:"id"`
	Points   [][]float64 `json:"points"`
}

// DB is the global Postgres database
var DB *sql.DB

// Queries
var upsetRoute = "INSERT INTO drivers_trips(driver_id, points) VALUES ($1, $2) ON CONFLICT (driver_id) DO UPDATE SET points = $2"
var updatePos = "UPDATE users SET current_lat = $2, current_lng = $3 WHERE id = $1"

// Conection info
const (
	host     = "ec2-54-228-170-125.eu-west-1.compute.amazonaws.com"
	port     = 5432
	user     = "gctgugggvhltaw"
	password = "f832bb546fa4574d3b49c28da0e7c8a007154de81aeaa82a14b62836fa11e929"
	dbname   = "dgobk4k3st1m1"
)

var upgrader = websocket.Upgrader{}
var err error

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname)

	// Prepare the database
	DB, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err.Error())
	}
	defer DB.Close()
	err = DB.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("DB connected!")

	http.HandleFunc("/ws", updateRoute)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	log.Println(port)
	srv := &http.Server{
		Addr: "0.0.0.0:" + port,
	}
	log.Fatal(srv.ListenAndServe())
}

func updateRoute(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()
	var route Route
	for {
		err = ws.ReadJSON(&route)
		if err != nil {
			log.Println("read:", err)
			break
		}
		_, err = DB.Exec(upsetRoute, route.DriverID, pq.Array(route.Points))
		if err != nil {
			panic(err.Error())
		}
		_, err = DB.Exec(updatePos, route.DriverID, route.Points[0][0], route.Points[0][1])
		if err != nil {
			panic(err.Error())
		}
		err = ws.WriteMessage(1, []byte("Success"))
		if err != nil {
			log.Println("write", err)
			break
		}
	}
}
