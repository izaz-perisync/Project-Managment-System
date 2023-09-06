package handler

import (
	"encoding/json"
	"io"
	"log"

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
	price := r.FormValue("Price")
	stock := r.FormValue("ProductCount")

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
		ProductName:  productName,
		Description:  description,
		Brand:        brand,
		Category:     category,
		FileData:     fileData,
		FileType:     fileType,
		Price:        price,
		ProductCount: stock,
	}

	err = service.AddProduct(tokenString, body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJson(w, http.StatusCreated, "product-added")
}

func HandlerUpdateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	tokenString := getTokenStringFromRequest(r)

	// err := r.ParseMultipartForm(32 << 20)
	// if err != nil {
	// 	writeJson(w, http.StatusBadRequest, "failed to parse data")
	// }

	// productId := r.FormValue("productId")
	// productName := r.FormValue("ProductName")
	// description := r.FormValue("Description")
	// brand := r.FormValue("Brand")
	// category := r.FormValue("Category")
	// fileType := r.FormValue("FileType")

	// file, _, err := r.FormFile("file")
	// if err != nil && err != http.ErrMissingFile {
	// 	writeJson(w, http.StatusInternalServerError, "failed to upload files")
	// }
	// var data []byte
	// if file != nil {
	// 	defer file.Close()
	// 	data, err = io.ReadAll(file)
	// 	if err != nil {
	// 		writeJson(w, http.StatusInternalServerError, "failed to read the file ")
	// 	}
	// }
	productId := r.URL.Query().Get("productId")

	prodId, err := strconv.ParseInt(productId, 0, 64)
	if err != nil {
		log.Println("111")
		writeJson(w, http.StatusBadRequest, "failed to parse")
		return
	}
	var body model.UpdateProduct
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	err = service.UpdateProduct(tokenString, body, prodId)
	if err != nil {

		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJson(w, http.StatusOK, "product updated")

}

func HandlerListProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var decoder = schema.NewDecoder()
	var body model.FilterByProductId
	err := decoder.Decode(&body, r.URL.Query())
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}

	ProductsList, err := service.ProductList(body)
	if err != nil {

		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	if ProductsList.TotalCount != 0 {
		json.NewEncoder(w).Encode(ProductsList)
		return
	}

	writeJson(w, http.StatusNoContent, ProductsList)

}

func HandlerFilterProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var body model.FilterProduct
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "failed to parse")
	}
	list, err := service.FilterProduct(body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	if list.TotalCount != 0 {
		json.NewEncoder(w).Encode(list)
		return
	}
	writeJson(w, http.StatusNoContent, list)
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
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	if msg == "delete success" {
		writeJson(w, http.StatusOK, msg)
		return
	}
	writeJson(w, http.StatusNoContent, msg)

}

func HandlerAddAssets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	tokenString := getTokenStringFromRequest(r)

	err := r.ParseMultipartForm(32 << 20) // Set an appropriate form size limit
	if err != nil {
		writeJson(w, http.StatusBadRequest, "unable to parse form data")
		return
	}

	fileType := r.FormValue("FileType")
	productId := r.FormValue("productId")
	file, _, err := r.FormFile("file")
	if err != nil {
		writeJson(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	id, err := strconv.ParseInt(productId, 0, 64)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "malformed request")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "reading file error")
		return
	}

	err = service.AddAssets(data, tokenString, fileType, id)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJson(w, http.StatusCreated, "assets-added")
}

func HandlerUpdateAsset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	tokenString := getTokenStringFromRequest(r)
	err := r.ParseMultipartForm(32 << 20) // Set an appropriate form size limit
	if err != nil {
		writeJson(w, http.StatusBadRequest, "unable to parse form data")
		return
	}

	fileType := r.FormValue("FileType")
	assetId := r.FormValue("assetId")
	file, _, err := r.FormFile("file")
	if err != nil {
		writeJson(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	id, err := strconv.ParseInt(assetId, 0, 64)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "malformed request")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "reading file error")
		return
	}

	body := model.UpdateAsset{
		AssetId:  int(id),
		FileDate: data,
		FileType: fileType,
	}

	err = service.UpdateAsset(body, tokenString)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJson(w, http.StatusCreated, "assets-added")
}

func HandleGetProductData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	product := r.URL.Query().Get("productId")
	id, err := strconv.ParseInt(product, 0, 64)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "error reading body")
		return
	}
	productData, err := service.GetProduct(id)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	if productData.Price != 0 {

		json.NewEncoder(w).Encode(productData)
		return
	}
	writeJson(w, http.StatusNoContent, nil)

}

func HandlerDeleteAsset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	product := r.URL.Query().Get("assetId")
	id, err := strconv.ParseInt(product, 0, 64)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "error reading body")
		return
	}
	w.Header().Set("content-type", "application/json")
	tokenString := getTokenStringFromRequest(r)

	err = service.DeleteAsset(tokenString, id)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJson(w, http.StatusOK, "asset-deleted")
}

func HandleAddToCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	tokenString := getTokenStringFromRequest(r)

	var decoder = schema.NewDecoder()
	var body model.FilterProduct
	err := decoder.Decode(&body, r.URL.Query())
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	err = service.AddTocart(body, tokenString)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJson(w, http.StatusCreated, "Product-added-to-Cart")

}

func HandlerDeleteCart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	tokenString := getTokenStringFromRequest(r)
	cartId := r.URL.Query().Get("cartId")

	id, err := strconv.ParseInt(cartId, 0, 64)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	err = service.RemoveCart(id, tokenString)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJson(w, http.StatusOK, "Cart-deleted")

}

func HandlerUpdateQuantity(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	tokenString := getTokenStringFromRequest(r)
	var decoder = schema.NewDecoder()
	var body model.FilterProduct
	err := decoder.Decode(&body, r.URL.Query())
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	err = service.UpdateQuantity(tokenString, body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	// json.NewEncoder(w).Encode("quantity updated")
	writeJson(w, http.StatusOK, "quantity updated")

}

func HandlerPlaceOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	tokenString := getTokenStringFromRequest(r)
	var decoder = schema.NewDecoder()
	var body model.FilterProduct
	err := decoder.Decode(&body, r.URL.Query())
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	err = service.PlaceOrder(tokenString, body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJson(w, http.StatusOK, "OrderPlaced")
}

func HandlerOrderList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	tokenString := getTokenStringFromRequest(r)
	list, err := service.OrderList(tokenString)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	if list.TotalCount != 0 {
		json.NewEncoder(w).Encode(list)
		return
	}
	writeJson(w, http.StatusNoContent, nil)
}

func HandlerOrderdetails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	orderId := r.URL.Query().Get("orderId")

	id, err := strconv.ParseInt(orderId, 0, 64)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	list, err := service.OrderDetails(id)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	if list != nil {
		json.NewEncoder(w).Encode(list)
		return
	}
	writeJson(w, http.StatusNoContent, nil)

}

func HandlerCartList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	tokenString := getTokenStringFromRequest(r)
	var decoder = schema.NewDecoder()
	var body model.FilterByProductId
	err := decoder.Decode(&body, r.URL.Query())
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	list, err := service.CartList(tokenString, body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	if list.TotalCount != 0 {
		json.NewEncoder(w).Encode(list)
		return
	}
	writeJson(w, http.StatusNoContent, nil)
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
