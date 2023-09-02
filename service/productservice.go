package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/perisynctechnologies/pms/model"
)

func AddProduct(token string, body model.AddProduct) (string, error) {
	// Validate the required fields
	if err := body.Validate(); err != nil {
		return "", err
	}

	id, err := VaildSign(token)
	if err != nil {
		return "", err
	}

	// Check if the product already exists based on your business logic
	find := MatchedProducts(body)
	if !find {
		var productid int

		// Insert product information into the database
		err = db.QueryRow(`insert into product
		 (userid, product_name, description, brand, category, created_at, updated_at) 
		 values ($1, $2, $3, $4, $5, $6, $7) returning product_id`,
			id, body.ProductName, body.Description, body.Brand, body.Category, time.Now(), nil).Scan(&productid)
		if err != nil {
			log.Println("here2")
			return "product not added", err
		}

		dirPath := fmt.Sprintf("./files/%d", productid)
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return "", err
		}

		if len(body.FileData) > 0 {

			filePath := filepath.Join("productfiles", body.ProductName+"."+body.FileType)
			err := os.WriteFile(filePath, body.FileData, 0644)
			if err != nil {
				return "product-added", err
			}
		}

		return "product-added", nil
	}

	return "product-already-exists", nil
}

func AddAsset(data []byte, token string, filetype string, product int64) (string, error) {
	id, err := VaildSign(token)
	if err != nil {
		return "", err
	}
	var userexsits, productexsits int
	err = db.QueryRow(`select userid,product_id from product where userid=$1 and product_id=$2`, id, product).Scan(&userexsits, &productexsits)
	if err != nil {
		return "", fmt.Errorf("product not found")
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	uuid := uuid.New().String()

	userFolderPath := filepath.Join("files", strconv.Itoa(int(product)))
	filepath := filepath.Join(userFolderPath, uuid+"."+filetype)

	outFile, err := os.Create(filepath)
	if err != nil {

		return "", err
	}
	defer outFile.Close()
	_, err = outFile.Write(data)
	if err != nil {
		return "", err
	}
	_, err = db.Exec(`insert into product_assets (productid,file_type,file_path,added_at,userid) values ($1,$2,$3,$4,$5)`, product, filetype, homeDir+filepath, time.Now(), id)
	if err != nil {
		return "", err
	}
	return "asset-added", nil

}

func UpdateProduct(token string, productId int64, body model.AddProduct) (*model.AddProduct, error) {
	id, err := VaildSign(token)
	if err != nil {
		return nil, err
	}

	result, err := db.Exec(`update product set product_name=$1,
	description=$2,brand=$3,category=$4,updated_At=$5 where userid=$6 and product_id=$7`, body.ProductName, body.Description, body.Brand, body.Category, time.Now(), id, productId)
	if err != nil {
		log.Println("e1")
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("error getting rows affected:", err)
		return nil, err
	}

	if rowsAffected == 0 {

		return nil, fmt.Errorf("product not found")
	}
	return &body, nil
}

func VaildSign(tokenString string) (*int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("my_secret_key"), nil
	})

	if err != nil {
		return nil, errors.New("error in parser")
	}

	if !token.Valid {
		return nil, errors.New("JWT token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims type in JWT token")
	}

	userIdFloat, ok := claims["userId"].(float64)
	if !ok {
		return nil, errors.New("userId not found in JWT claims")
	}

	userId := int(userIdFloat)

	return &userId, nil
}

func MatchedProducts(body model.AddProduct) bool {

	var count int
	err := db.QueryRow(`select count(*) from product where product_name=$1
 and description=$2 and category=$3 and brand=$4`,
		body.ProductName, body.Description, body.Category, body.Brand).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

func ProductList(body model.FilterByProductId) (*model.ListProducts, error) {

	query := `select product_id,product_name,description,brand,category from product `

	if body.ProductId != 0 {
		query += " where " + " product_id = " + strconv.Itoa(body.ProductId)

	}
	if body.Size == 0 {
		body.Size = 10
	}
	offset := 0
	if body.Page != 0 {

		if body.Page > 1 {
			offset = body.Size * (body.Page - 1)
		}
	}
	query += fmt.Sprintf(" ORDER BY product_id DESC OFFSET %d LIMIT %d ", offset, body.Size)

	var data model.ListProducts

	data.TotalCount = 0
	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var list model.ProductDetails

		err := rows.Scan(&list.ProductId, &list.ProductName, &list.Description, &list.Brand, &list.Category)

		if err != nil {
			return nil, err
		}
		fmt.Println(list)
		row2, err := db.Query(`select asset_id,file_path,file_type,added_at from product_assets where productid=$1`, list.ProductId)
		if err != nil {
			return nil, err
		}
		defer row2.Close()
		for row2.Next() {
			var assetlist model.Assets
			err := row2.Scan(&assetlist.AssetId, &assetlist.FilePath, &assetlist.AssetType, &assetlist.Added_at)
			if err != nil {
				return nil, err
			}

			list.Assets = append(list.Assets, assetlist)

		}

		data.TotalCount++

		data.ProductList = append(data.ProductList, list)

	}
	return &data, nil
}

func DeleteProduct(token string, prodId int64) (string, error) {
	id, err := VaildSign(token)
	if err != nil {
		return "", err
	}
	fmt.Println("prodid ", prodId)
	result, err := db.Exec(`DELETE 
	FROM product
	WHERE product_id = $1 and userid=$2
	`, prodId, id)
	if err != nil {

		return "", err
	}

	fmt.Println("userid", *id, "prodid", prodId)

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("error getting rows affected:", err)
		return "", err
	}

	_, err = db.Exec(`delete from product_assets where productid=$1 and userid=$2`, prodId, id)
	if err != nil {
		return "", err
	}

	if rowsAffected == 0 {

		return "not deleted", err
	}
	dirPath := fmt.Sprintf("./files/%d", prodId)
	err = os.RemoveAll(dirPath)
	if err != nil {
		return "", err
	}

	return "delete success", nil
}
