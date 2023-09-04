package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/perisynctechnologies/pms/handler"
	"github.com/perisynctechnologies/pms/service"
)

func main() {
	log.Println("connecting to  db...")
	constr := "postgres://izaz:aHr4SLa9dAGY@postgresql-141573-0.cloudclusters.net:10034/trainer?sslmode=disable"
	Db, err := sql.Open("postgres", constr)

	if err != nil {
		log.Fatal(err)
		return
	}

	defer Db.Close()

	if err = Db.Ping(); err != nil {
		log.Fatal(err)
		return
	}
	service.IntilizeDb(Db)
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	user := api.PathPrefix("/account").Subrouter()
	user.HandleFunc("/register", handler.HandlerRegister).Methods(http.MethodPost)
	user.HandleFunc("/login", handler.HandlerLogin).Methods(http.MethodPost)

	product := api.PathPrefix("/product").Subrouter()
	product.HandleFunc("/add", handler.HandlerAddProduct).Methods(http.MethodPost)
	product.HandleFunc("/add_asset", handler.HandlerAddAssets).Methods(http.MethodPost)
	product.HandleFunc("/update", handler.HandlerUpdateProduct).Methods(http.MethodPut)
	product.HandleFunc("/list", handler.HandlerListProducts).Methods(http.MethodGet)
	product.HandleFunc("/delete", handler.HandlerDeleteProduct).Methods(http.MethodDelete)
	product.HandleFunc("/update_asset", handler.HandlerUpdateAsset).Methods(http.MethodPut)
	product.HandleFunc("/details", handler.HandleGetProductData).Methods(http.MethodGet)
	product.HandleFunc("/delete_asset", handler.HandlerDeleteAsset).Methods(http.MethodDelete)

	log.Fatal(http.ListenAndServe(":3000", r))

}
