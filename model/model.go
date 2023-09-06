package model

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Register struct {
	FirstName    string `json:"firstName" `
	MiddleName   string `json:"middleName"`
	LastName     string `json:"lastName"`
	MobileNumber int    `json:"mobileNumber"`
	Email        string `json:"email" `
	Password     string `json:"password" `
}
type Login struct {
	Email    string `json:"email" `
	Password string `json:"password" `
}
type LogUser struct {
	FirstName  string    `json:"firstName" `
	MiddleName string    `json:"middleName"`
	LastName   string    `json:"lastName"`
	Email      string    `json:"email" `
	CreatedAt  time.Time `json:"createdAt"`
	Token      string    `json:"token"`
}
type Claims struct {
	UserId int `json:"userId"`
	jwt.RegisteredClaims
}

type AddProduct struct {
	ProductName  string `json:"productName"`
	Description  string `json:"description"`
	Brand        string `json:"brand"`
	Category     string `json:"category"`
	FileData     []byte `json:"fileData"`
	FileType     string `json:"fileType"`
	FileDetails  string `json:"fileDeatails"`
	Price        string `json:"price"`
	ProductCount string `json:"stock"`
}

type ListedProducts struct {
	ProductName string `json:"productName"`
	Description string `json:"description"`
	Brand       string `json:"brand"`
	Category    string `json:"category"`
	Price       int    `json:"price"`
	Stock       int    `json:"stock"`
}

type UpdateProduct struct {
	ProductName string `json:"productName"`
	Description string `json:"description"`
	Brand       string `json:"brand"`
	Category    string `json:"category"`
	Price       int    `json:"price"`
	Stock       int    `json:"stock"`
}

type Assets struct {
	AssetId   int       `json:"assetId"`
	AssetType string    `json:"assetType"`
	FilePath  string    `json:"filePath"`
	Added_at  time.Time `json:"added_At"`
}
type ListWithoutStock struct {
	ProductName string `json:"productName"`
	Description string `json:"description"`
	Brand       string `json:"brand"`
	Category    string `json:"category"`
	Price       int    `json:"price"`
}

type ProductDetails struct {
	ProductId int `json:"productId"`
	ListedProducts
	Assets []Assets
}

type Singleproduct struct {
	ListedProducts
	Assets []Assets
}
type ProductData struct {
	ListWithoutStock
	Assets []Assets
}

type ListProducts struct {
	TotalCount  int `json:"totalCount"`
	ProductList []ProductDetails
}

type FilterByProductId struct {
	ProductId int    `schema:"productId"`
	Page      int    `schema:"page"`
	Size      int    `schema:"size"`
	PriceMin  int    `schema:"PriceMin"`
	PriceMax  int    `schema:"PriceMax"`
	Sort      string `schema:"sort"`
}

type FilterProduct struct {
	ProductName string `schema:"productName"`
	Brand       string `schema:"brand"`
	Category    string `schema:"category"`
	PriceMin    int    `schema:"priceMin"`
	PriceMax    int    `schema:"priceMax"`
	Sort        string `schema:"sort"`
	Page        int    `schema:"page"`
	Size        int    `schema:"size"`
	SortColumn  string `schema:"sortColumn"`
	SortOrder   string `schema:"sortOrder"`
}

type UpdateAsset struct {
	AssetId  int    `json:"assetId"`
	FileDate []byte `json:"fileData"`
	FileType string `json:"fileType"`
}

type CartDetails struct {
	CartId int `json:"cartid"`
	ListedProducts
	ProductCount int `json:"productCount"`
	Assets       []Assets
}

type CartList struct {
	TotalCount int `json:"totalCount"`
	CartData   []CartDetails
}

type OrderDetails struct {
	OrderId int `json:"orderId"`
	ListedProducts
	Assets []Assets
}

type OrderData struct {
	TotalCount  int `json:"totalCount"`
	ProductInfo []ProductData
}

type OrderList struct {
	TotalCount   int `json:"totalCount"`
	OrderDetails []OrderDetails
}

// if productName == "" || description == "" || brand == "" || category == "" {
// 	writeJson(w, http.StatusBadRequest, "all fields are required")
// 	return
// }

func (u *UpdateAsset) Validate() error {
	if strconv.Itoa(u.AssetId) == "" {
		return fmt.Errorf("assetId is empty")
	}
	if u.FileDate == nil {
		return fmt.Errorf("file is empty")
	}
	if u.FileType == "" {
		return fmt.Errorf("file type is empty")
	}
	return nil
}

func (u *Register) Validate() error {

	if u.FirstName == "" {
		return fmt.Errorf("first name is empty")
	}
	if u.Email == "" {
		return fmt.Errorf("enter Email")
	}

	if u.Password == "" || len(u.Password) < 8 {
		return fmt.Errorf("enter valid Password")
	}

	isValid := isValidEmail(u.Email)

	if !isValid {

		fmt.Println("Email not valid")
		return fmt.Errorf("email not valid")
	}
	if u.MobileNumber == 0 {
		fmt.Println("error")
		return nil
	} else if len(strconv.Itoa(u.MobileNumber)) < 10 {
		return fmt.Errorf("not a valid mobile number")
	}
	return nil
}

func (l *Login) Validate() error {
	if l.Email == "" {
		return fmt.Errorf("enter Email")
	}
	if len(l.Password) < 8 {
		return fmt.Errorf("enter valid password")
	}
	isValid := isValidEmail(l.Email)

	if !isValid {
		return fmt.Errorf("email not valid")
	}
	return nil
}

func (a *AddProduct) Validate() error {
	if a.ProductName == "" {
		return fmt.Errorf("product name is empty")
	}
	if a.Brand == "" {
		return fmt.Errorf("brand is empty")
	}
	if a.Category == "" {
		return fmt.Errorf("category is empty")
	}
	if a.Description == "" {
		return fmt.Errorf("description is empty")
	}
	if a.FileData != nil {
		if a.FileType == "" {
			return fmt.Errorf("file type is empty")
		}
	}
	if a.Price == "" {
		return fmt.Errorf("price is empty")
	}
	if a.ProductCount == "" {
		return fmt.Errorf("stock is empty")
	}
	return nil
}
func (a *UpdateProduct) Validate() error {

	if a.ProductName == "" {
		return fmt.Errorf("product name is empty")
	}
	if a.Brand == "" {
		return fmt.Errorf("brand is empty")
	}
	if a.Category == "" {
		return fmt.Errorf("category is empty")
	}
	if a.Description == "" {
		return fmt.Errorf("description is empty")
	}
	if strconv.Itoa(a.Price) == "" {
		return fmt.Errorf("price is empty")
	}
	if strconv.Itoa(a.Stock) == "" {
		return fmt.Errorf("stock is empty")
	}

	return nil
}
func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}
