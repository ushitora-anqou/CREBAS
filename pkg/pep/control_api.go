package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// json: ~ をパラメータに付与すると、jsonエンコード時にパラメータ名を指定することができます。
// また、omitemptyを付与するとパラメータが空のときに、jsonのパラメータから消すことができます。
// これはクライアントアプリと仕様を統一する必要があります。
type ItemParams struct {
	Id           string    `json:"id"`
	JanCode      string    `json:"jan_code,omitempty"`
	ItemName     string    `json:"item_name,omitempty"`
	Price        int       `json:"price,omitempty"`
	CategoryId   int       `json:"category_id,omitempty"`
	SeriesId     int       `json:"series_id,omitempty"`
	Stock        int       `json:"stock,omitempty"`
	Discontinued bool      `json:"discontinued"`
	ReleaseDate  time.Time `json:"release_date,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    time.Time `json:"deleted_at"`
}

// ポインタ型でitemsを定義します。今回はこのグローバル変数【配列】がデータベースの役割をします
var items []*ItemParams

func rootPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the Go Api Server")
	fmt.Println("Root endpoint is hooked!")
}

func getAllApps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps.GetAllAppInfos())
}

func fetchSingleItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	key := vars["id"]

	for _, item := range items {
		if item.Id == key {
			json.NewEncoder(w).Encode(item)
		}
	}
}

func createItem(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var item ItemParams
	if err := json.Unmarshal(reqBody, &item); err != nil {
		log.Fatal(err)
	}

	items = append(items, &item)
	json.NewEncoder(w).Encode(item)
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	for index, item := range items {
		if item.Id == id {
			items = append(items[:index], items[index+1:]...)
		}
	}
}

func updateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	reqBody, _ := ioutil.ReadAll(r.Body)
	var updateItem ItemParams
	if err := json.Unmarshal(reqBody, &updateItem); err != nil {
		log.Fatal(err)
	}

	for index, item := range items {
		if item.Id == id {
			items[index] = &ItemParams{
				Id:           item.Id,
				JanCode:      updateItem.JanCode,
				ItemName:     updateItem.ItemName,
				Price:        updateItem.Price,
				CategoryId:   updateItem.CategoryId,
				SeriesId:     updateItem.SeriesId,
				Stock:        updateItem.Stock,
				Discontinued: updateItem.Discontinued,
				ReleaseDate:  updateItem.ReleaseDate,
				CreatedAt:    item.CreatedAt,
				UpdatedAt:    updateItem.UpdatedAt,
				DeletedAt:    item.DeletedAt,
			}
		}
	}
}

func getAllPkgs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pkgs.GetAll())
}

func StartWebServer() error {
	fmt.Println("Rest API with Mux Routers")
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/pkgs", getAllPkgs).Methods("GET")
	router.HandleFunc("/apps", getAllApps).Methods("GET")
	router.HandleFunc("/", rootPage)
	router.HandleFunc("/item/{id}", fetchSingleItem).Methods("GET")

	router.HandleFunc("/item", createItem).Methods("POST")
	router.HandleFunc("/item/{id}", deleteItem).Methods("DELETE")
	router.HandleFunc("/item/{id}", updateItem).Methods("PUT")

	return http.ListenAndServe(fmt.Sprintf(":%d", 8080), router)
}

// モックデータを初期値として読み込みます
func init() {
}
