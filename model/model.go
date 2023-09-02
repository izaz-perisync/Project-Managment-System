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
	ProductName string `json:"productName"`
	Description string `json:"description"`
	Brand       string `json:"brand"`
	Category    string `json:"category"`
	FileData    []byte `json:"fileData"`
	FileType    string `json:"fileType"`
}
type Assets struct {
	AssetId   int       `json:"assetId"`
	AssetType string    `json:"assetType"`
	FilePath  string    `json:"filePath"`
	Added_at  time.Time `json:"added_At"`
}

type ProductDetails struct {
	ProductId int `json:"productId"`
	AddProduct
	Assets []Assets
}

type ListProducts struct {
	TotalCount  int `json:"totalCount"`
	ProductList []ProductDetails
}

type FilterByProductId struct {
	ProductId int `schema:"productId"`
	Page      int `schema:"page"`
	Size      int `schema:"size"`
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

func (p *AddProduct) Validate() error {
	if p.ProductName == "" {
		return fmt.Errorf("product name is empty")
	}
	if p.Description == "" {
		return fmt.Errorf("product description is empty")
	}
	if p.Category == "" {
		return fmt.Errorf(" category is empty")
	}
	if p.Brand == "" {
		return fmt.Errorf("brand  is empty")
	}
	return nil
}

func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}
