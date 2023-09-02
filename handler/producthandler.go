package handler

import (
	"encoding/json"
	"io"

	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/schema"
	"github.com/perisynctechnologies/pms/model"
	"github.com/perisynctechnologies/pms/service"
)

func HandlerAddProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := getTokenStringFromRequest(r)

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "failed to parse form data")
		return
	}

	productName := r.FormValue("ProductName")
	description := r.FormValue("Description")
	brand := r.FormValue("Brand")
	category := r.FormValue("Category")
	fileType := r.FormValue("FileType")

	if productName == "" || description == "" || brand == "" || category == "" {
		writeJson(w, http.StatusBadRequest, "all fields are required")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil && err != http.ErrMissingFile {
		writeJson(w, http.StatusInternalServerError, "failed to process file upload")
		return
	}

	var fileData []byte
	if file != nil {
		defer file.Close()

		fileData, err = io.ReadAll(file)
		if err != nil {
			writeJson(w, http.StatusInternalServerError, "failed to read file data")
			return
		}
	}

	body := model.AddProduct{
		ProductName: productName,
		Description: description,
		Brand:       brand,
		Category:    category,
		FileData:    fileData,
		FileType:    fileType,
	}

	msg, err := service.AddProduct(tokenString, body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}

	if msg == "product-added" {
		writeJson(w, http.StatusCreated, "product-added")
		return
	}

	writeJson(w, http.StatusBadRequest, msg)
}

func HandlerUpdateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	tokenString := getTokenStringFromRequest(r)
	product := r.URL.Query().Get("productId")
	id, err := strconv.ParseInt(product, 0, 64)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "error reading body")
		return
	}
	var body model.AddProduct
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "malformed request")
		return
	}
	updatedProduct, err := service.UpdateProduct(tokenString, id, body)
	if err != nil {

		writeJson(w, http.StatusBadRequest, "update failure")
		return
	}
	if updatedProduct != nil {

		json.NewEncoder(w).Encode(updatedProduct)
		return
	}

}

func HandlerListProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var decoder = schema.NewDecoder()
	var body model.FilterByProductId
	err := decoder.Decode(&body, r.URL.Query())
	if err != nil {
		writeJson(w, http.StatusBadRequest, err)
		return
	}

	ProductsList, err := service.ProductList(body)
	if err != nil {

		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	if ProductsList != nil {
		json.NewEncoder(w).Encode(ProductsList)
		return
	}
	writeJson(w, http.StatusNoContent, ProductsList)

}

func HandlerDeleteProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	tokenString := getTokenStringFromRequest(r)
	product := r.URL.Query().Get("productId")
	id, err := strconv.ParseInt(product, 0, 64)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "error reading body")
		return
	}

	msg, err := service.DeleteProduct(tokenString, id)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err)
		return
	}
	if msg != "" {
		writeJson(w, http.StatusOK, msg)
		return
	}
	writeJson(w, http.StatusNoContent, msg)

}

func HandlerAddAssets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	tokenString := getTokenStringFromRequest(r)

	// Parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // Set an appropriate form size limit
	if err != nil {
		writeJson(w, http.StatusBadRequest, "unable to parse form data")
		return
	}

	fileType := r.FormValue("FileType")
	productID := r.FormValue("productId")
	file, _, err := r.FormFile("file")
	if err != nil {
		writeJson(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	id, err := strconv.ParseInt(productID, 0, 64)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "malformed request")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "reading file error")
		return
	}

	msg, err := service.AddAsset(data, tokenString, fileType, id)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	if msg != "" {
		writeJson(w, http.StatusCreated, msg)
		return
	}
	writeJson(w, http.StatusBadRequest, msg)
}

func getTokenStringFromRequest(r *http.Request) string {
	// Retrieve the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	//our auth header looks like this Authorization: Bearer <token-value>
	// Check if the header has the Bearer token format
	if strings.HasPrefix(authHeader, "Bearer ") {

		//use of authHeader[len("Bearer "):] it removes the bearer tag of length of 7 and return the token string "Bearer abcdef123456"
		return authHeader[len("Bearer "):]
	}

	return ""
}
