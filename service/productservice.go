package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/lib/pq"
	"github.com/perisynctechnologies/pms/model"
)

func AddProduct(token string, body model.AddProduct) error {

	id, err := VaildSign(token)
	if err != nil {
		return fmt.Errorf("invalid token")
	}

	if err := body.Validate(); err != nil {
		return err
	}

	find := MatchedProducts(body, id)

	log.Println(find)
	if !find {
		var productid int
		price, err := strconv.ParseInt(body.Price, 0, 64)
		if err != nil {
			return err
		}

		stock, err := strconv.ParseInt(body.ProductCount, 0, 64)
		if err != nil {
			return err
		}

		err = db.QueryRow(`insert into product
		 (userid, product_name, description, brand, category, created_at, updated_at,price,stock) 
		 values ($1, $2, $3, $4, $5, $6, $7,$8,$9) returning product_id`,
			id, body.ProductName, body.Description, body.Brand, body.Category, time.Now(), nil, price, stock).Scan(&productid)
		if err != nil {

			return fmt.Errorf("product not added")
		}

		// dirPath := fmt.Sprintf("./files/%d", productid)
		// err = os.MkdirAll(dirPath, os.ModePerm)
		// if err != nil {
		// 	return err
		// }

		if len(body.FileData) > 0 {
			dirPath := fmt.Sprintf("./productfiles/%d", productid)
			err = os.MkdirAll(dirPath, os.ModePerm)
			if err != nil {
				return err
			}

			var asset_id int
			err = db.QueryRow(`insert into product_assets (productid,file_type,added_at,userid) values ($1,$2,$3,$4) returning asset_id`, productid, body.FileType, time.Now(), id).Scan(&asset_id)
			if err != nil {
				return err
			}
			userFolderPath := filepath.Join("productfiles", strconv.Itoa(int(productid)))
			filePath := filepath.Join(userFolderPath, strconv.Itoa(asset_id)+"."+body.FileType)
			outFile, err := os.Create(filePath)
			if err != nil {

				return err
			}
			_, err = db.Exec(`update  product_assets  set file_path=$1 where asset_id=$2`, filePath, asset_id)
			if err != nil {
				return err
			}

			defer outFile.Close()
			_, err = outFile.Write(body.FileData)
			if err != nil {

				return err
			}
			// _, err = db.Exec(`update  product set filename=$1 where product_id=$2`, strconv.Itoa(productid)+"."+body.FileType, productid)
			// if err != nil {

			// 	return err
			// }

		}

		return nil
	}

	return fmt.Errorf("product-already-exsits")
}

func AddAssets(data []byte, token string, filetype string, product int64) error {
	id, err := VaildSign(token)
	if err != nil {
		return err
	}
	var userexsits, productexsits int
	err = db.QueryRow(`select userid,product_id from product where userid=$1 and product_id=$2`, id, product).Scan(&userexsits, &productexsits)
	if err != nil {
		return fmt.Errorf("product not found")
	}

	var asset_id int
	err = db.QueryRow(`insert into product_assets (productid,file_type,added_at,userid) values ($1,$2,$3,$4) returning asset_id`, product, filetype, time.Now(), id).Scan(&asset_id)
	if err != nil {
		return err
	}

	dirPath := fmt.Sprintf("./productfiles/%d", product)
	_, err = os.Stat(dirPath)
	if err != nil {
		dirPath = fmt.Sprintf("./productfiles/%d", product)
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	filepath := filepath.Join(dirPath, strconv.Itoa(asset_id)+"."+filetype)
	_, err = db.Exec(`update  product_assets  set file_path=$1 where asset_id=$2`, filepath, asset_id)
	if err != nil {
		return err
	}

	outFile, err := os.Create(filepath)
	if err != nil {

		return err
	}
	defer outFile.Close()
	_, err = outFile.Write(data)
	if err != nil {
		return err
	}
	return nil

}

func UpdateProduct(token string, body model.UpdateProduct, prodId int64) error {
	id, err := VaildSign(token)
	if err != nil {
		return err
	}
	if err := body.Validate(); err != nil {
		return err
	}
	// price, err := strconv.ParseInt(body.Price, 0, 64)
	// if err != nil {
	// 	return err
	// }

	result, err := db.Exec(`update product set product_name=$1,
	description=$2,brand=$3,category=$4,updated_At=$5,price=$6,stock=$7 where userid=$8 and product_id=$9`, body.ProductName, body.Description, body.Brand, body.Category, time.Now(), body.Price, body.Stock, id, prodId)
	if err != nil {

		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("error getting rows affected:", err)
		return err
	}

	if rowsAffected == 0 {

		return fmt.Errorf("product not found")
	}

	// if len(body.FileData) > 0 {

	// 	dirPath := fmt.Sprintf("./productfiles/%d", body.ProductId)
	// 	if err := os.RemoveAll(dirPath); err != nil {
	// 		return nil, err
	// 	}

	// 	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
	// 		return nil, err
	// 	}

	// 	userFolderPath := filepath.Join("productfiles", strconv.Itoa(int(body.ProductId)))
	// 	filePath := filepath.Join(userFolderPath, strconv.Itoa(body.ProductId)+"."+body.FileType)
	// 	outFile, err := os.Create(filePath)
	// 	if err != nil {

	// 		return nil, err
	// 	}
	// 	defer outFile.Close()
	// 	_, err = outFile.Write(body.FileData)
	// 	if err != nil {

	// 		return nil, err
	// 	}

	// }

	return nil
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

func MatchedProducts(body model.AddProduct, id *int) bool {

	var count int
	err := db.QueryRow(`select count(*) from product where 
  userid=$1 and product_name=$2`,
		id, body.ProductName).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}

func ProductList(body model.FilterByProductId) (*model.ListProducts, error) {

	query := `select product_id,product_name,description,brand,category,price,stock from product `

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
	if body.PriceMax == 0 && body.PriceMin == 0 && body.Sort == "" {

		query += fmt.Sprintf(" ORDER BY product_id DESC OFFSET %d LIMIT %d ", offset, body.Size)
	}

	if body.PriceMin != 0 && body.PriceMax != 0 && body.Sort == "" {
		query += fmt.Sprintf(" where price BETWEEN %d AND %d ORDER BY price DESC OFFSET %d LIMIT %d ", body.PriceMin, body.PriceMax, offset, body.Size)
	} else if body.PriceMin != 0 && body.Sort == "" {
		query += fmt.Sprintf(" where price >= %d ORDER BY price DESC OFFSET %d LIMIT %d  ", body.PriceMin, offset, body.Size)
	} else if body.PriceMax != 0 && body.Sort == "" {
		query += fmt.Sprintf(" where price <= %d ORDER BY price DESC OFFSET %d LIMIT %d", body.PriceMax, offset, body.Size)
	}
	if body.PriceMin != 0 && body.PriceMax != 0 && body.Sort != "" {
		query += fmt.Sprintf(" where price BETWEEN %d AND %d ORDER BY price %s OFFSET %d LIMIT %d ", body.PriceMin, body.PriceMax, body.Sort, offset, body.Size)
	} else if body.PriceMin != 0 && body.Sort != "" {
		query += fmt.Sprintf(" where price >= %d ORDER BY price %s OFFSET %d LIMIT %d  ", body.PriceMin, body.Sort, offset, body.Size)
	} else if body.PriceMax != 0 && body.Sort != "" {
		query += fmt.Sprintf(" where price <= %d ORDER BY price %s OFFSET %d LIMIT %d  ", body.PriceMin, body.Sort, offset, body.Size)
	}

	var data model.ListProducts

	data.TotalCount = 0
	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var list model.ProductDetails

		err := rows.Scan(&list.ProductId, &list.ProductName, &list.Description, &list.Brand, &list.Category, &list.Price, &list.Stock)

		if err != nil {
			return nil, err
		}
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

	result, err := db.Exec(`DELETE 
	FROM product
	WHERE product_id = $1 and userid=$2
	`, prodId, id)
	if err != nil {

		return "", fmt.Errorf("error in delete")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {

		return "", err
	}

	_, err = db.Exec(`delete from product_assets where productid=$1 and userid=$2`, prodId, id)
	if err != nil {
		return "", err
	}

	if rowsAffected == 0 {

		return "product not found", nil
	}

	path := fmt.Sprintf("./productfiles/%d", prodId)
	err = os.RemoveAll(path)
	if err != nil {
		return "", err
	}

	dirPath := fmt.Sprintf("./files/%d", prodId)
	err = os.RemoveAll(dirPath)
	if err != nil {
		return "", err
	}

	return "delete success", nil
}

func UpdateAsset(body model.UpdateAsset, token string) error {
	_, err := VaildSign(token)
	if err != nil {
		return err
	}

	if err := body.Validate(); err != nil {
		return err
	}

	var product_id int
	var file string
	err = db.QueryRow(`select productid,file_path from product_assets where asset_id =$1`, body.AssetId).Scan(&product_id, &file)
	if err != nil {
		return err
	}
	dirPath := file
	fmt.Println(dirPath)
	if err := os.Remove(dirPath); err != nil {
		return err
	}

	file_path := fmt.Sprintf("./productfiles/%d", product_id)

	filepath := filepath.Join(file_path, strconv.Itoa(body.AssetId)+"."+body.FileType)
	outFile, err := os.Create(filepath)
	if err != nil {

		return err
	}
	defer outFile.Close()
	_, err = outFile.Write(body.FileDate)
	if err != nil {
		return err
	}

	result, err := db.Exec(`update  product_assets  set file_path=$1 ,added_at=$2 where asset_id=$3`, filepath, time.Now(), body.AssetId)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {

		return err
	}
	if rowsAffected == 0 {

		return fmt.Errorf("product not found")
	}
	return nil
}

func GetProduct(id int64) (*model.Singleproduct, error) {
	query := `select product_name,description,brand,category,price from product where  product_id =$1  `
	rows, err := db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var data model.Singleproduct
	for rows.Next() {

		err := rows.Scan(&data.ProductName, &data.Description, &data.Brand, &data.Category, &data.Price)
		if err != nil {
			return nil, err
		}
		row2, err := db.Query(`select asset_id,file_path,file_type,added_at from product_assets where productid=$1`, id)
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
			data.Assets = append(data.Assets, assetlist)
			return &data, nil
		}

	}
	return nil, nil
}

func DeleteAsset(token string, asset_id int64) error {
	id, err := VaildSign(token)
	if err != nil {
		return err
	}
	var file string
	err = db.QueryRow(`select file_path from product_assets where asset_id =$1 and userid=$2`, asset_id, id).Scan(&file)
	if err != nil {
		return err
	}
	dirPath := file
	fmt.Println(dirPath)
	if err := os.Remove(dirPath); err != nil {
		return err
	}
	result, err := db.Exec(`DELETE 
	FROM product_assets
	WHERE asset_id = $1 and userid=$2
	`, asset_id, id)
	if err != nil {

		return fmt.Errorf("error in delete")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {

		return err
	}
	if rowsAffected == 0 {

		return fmt.Errorf("no asset found")
	}
	return nil

}

func FilterProduct(body model.FilterProduct) (*model.ListProducts, error) {

	query := `select product_id,product_name,description,brand,category,price,stock from product `
	if body.Size == 0 {
		body.Size = 10
	}
	offset := 0
	if body.Page != 0 {

		if body.Page > 1 {
			offset = body.Size * (body.Page - 1)
		}
	}
	var paramValues []interface{}

	var conditions []string

	if body.ProductName != "" {
		conditions = append(conditions, "product_name ILIKE $"+strconv.Itoa(len(paramValues)+1))
		paramValues = append(paramValues, "%"+body.ProductName+"%")

	}
	if body.Brand != "" {
		conditions = append(conditions, "brand ILIKE $"+strconv.Itoa(len(paramValues)+1))
		// paramValues = append(paramValues, body.Brand)
		paramValues = append(paramValues, "%"+body.Brand+"%")
	}
	if body.Category != "" {
		conditions = append(conditions, "category ILIKE $"+strconv.Itoa(len(paramValues)+1))
		// paramValues = append(paramValues, body.Category)
		paramValues = append(paramValues, "%"+body.Category+"%")
	}

	if body.PriceMin != 0 && body.PriceMax != 0 {
		conditions = append(conditions, "price BETWEEN $"+strconv.Itoa(len(paramValues)+1)+" AND $"+strconv.Itoa(len(paramValues)+2))
		paramValues = append(paramValues, body.PriceMin)
		paramValues = append(paramValues, body.PriceMax)
	} else if body.PriceMin != 0 {
		conditions = append(conditions, "price >= $"+strconv.Itoa(len(paramValues)+1))
		paramValues = append(paramValues, body.PriceMin)
	} else if body.PriceMax != 0 {
		conditions = append(conditions, "price <= $ "+strconv.Itoa(len(paramValues)+1))
		paramValues = append(paramValues, body.PriceMax)
	}
	fmt.Println("===", conditions)
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")

		log.Println(strings.Join(conditions, " AND "))
	}

	if body.SortColumn != "" {
		query += " ORDER BY " + body.SortColumn + " " + body.SortOrder
	} else {

		query += " ORDER BY product_id DESC"
	}

	query += fmt.Sprintf(" OFFSET %d LIMIT %d", offset, body.Size)

	var data model.ListProducts
	fmt.Println(query)
	data.TotalCount = 0
	rows, err := db.Query(query, paramValues...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var list model.ProductDetails

		err := rows.Scan(&list.ProductId, &list.ProductName, &list.Description, &list.Brand, &list.Category, &list.Price, &list.Stock)

		if err != nil {
			return nil, err
		}
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

func AddTocart(productId int64, token string) error {
	id, err := VaildSign(token)
	if err != nil {
		return err
	}
	var productexsits, stock int
	err = db.QueryRow(`select product_id,stock from product where product_id=$1`, productId).Scan(&productexsits, &stock)
	if err != nil {
		return errors.New("product not found")
	}

	if stock <= 0 {
		return errors.New("product-out-of-stock")
	}

	count, cartid := matched(productId, id)
	if count > 0 {

		_, err = db.Exec(`update cart set product_count=$1 where cartid=$2`, count+1, cartid)
		if err != nil {
			return err
		}
		return nil
	}

	_, err = db.Exec(`insert into cart (productid,product_count,userid) values ($1,$2,$3)`, productexsits, count+1, id)
	if err != nil {
		return err
	}

	return nil

}

func RemoveCart(cartId int64, token string) error {
	id, err := VaildSign(token)
	if err != nil {
		return err
	}
	result, err := db.Exec(`delete from cart where cartid=$1 and userid=$2`, cartId, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {

		return err
	}
	if rowsAffected == 0 {

		return fmt.Errorf("cart not found")
	}
	return nil
}

func CartList(token string, body model.FilterByProductId) (*model.CartList, error) {
	id, err := VaildSign(token)
	if err != nil {
		return nil, err
	}
	query := "select productid,product_count,cartid from cart where userid=$1"
	if body.Size == 0 {
		body.Size = 10
	}
	offset := 0
	if body.Page != 0 {

		if body.Page > 1 {
			offset = body.Size * (body.Page - 1)
		}
	}

	query += fmt.Sprintf(" ORDER BY cartid DESC OFFSET %d LIMIT %d", offset, body.Size)

	row, err := db.Query(query, id)

	fmt.Println(query)
	if err != nil {
		return nil, err
	}
	var list model.CartList
	list.TotalCount = 0
	defer row.Close()
	for row.Next() {
		var body model.CartDetails
		var product int
		err := row.Scan(&product, &body.ProductCount, &body.CartId)
		if err != nil {
			return nil, err
		}
		rows2, err := db.Query(`select product_name,description,brand,category,price,stock from product where product_id=$1`, product)
		if err != nil {
			return nil, err
		}
		defer rows2.Close()
		for rows2.Next() {
			err := rows2.Scan(&body.ProductName, &body.Description, &body.Brand, &body.Category, &body.Price, &body.Stock)
			if err != nil {
				return nil, err
			}
			row2, err := db.Query(`select asset_id,file_path,file_type,added_at from product_assets where productid=$1`, product)
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

				body.Assets = append(body.Assets, assetlist)

			}

		}
		list.TotalCount++
		list.CartData = append(list.CartData, body)
	}
	return &list, nil

}

// func PlaceOrder(token string) error {
// 	id, err := VaildSign(token)
// 	if err != nil {
// 		return err
// 	}
// 	var cartid, product []int
// 	err = db.QueryRow(`select cartid,productid from cart where userid =$1`, id).Scan(&cartid, &product)
// 	if err != nil {
// 		return errors.New("no item in cart")
// 	}
// 	fmt.Println(cartid, product)
// 	return nil
// }

func PlaceOrder(token string) error {
	id, err := VaildSign(token)
	if err != nil {
		return err
	}

	// Use db.Query to retrieve multiple rows
	rows, err := db.Query(`SELECT cartid, productid FROM cart WHERE userid = $1`, id)
	if err != nil {
		return err
	}
	defer rows.Close()

	var cartids []int
	var products []int

	// Iterate through the rows and append values to the arrays
	for rows.Next() {
		var cartid, productid int
		if err := rows.Scan(&cartid, &productid); err != nil {
			return err
		}
		cartids = append(cartids, cartid)
		products = append(products, productid)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return err
	}

	cartidsArray := pq.Array(cartids)
	productsArray := pq.Array(products)


	db.QueryRow(`select stock from product where product_id=$1`,)

	_, err = db.Exec(`INSERT INTO orders (orderlist, userid, productid) VALUES ($1, $2, $3)`, cartidsArray, id, productsArray)

	if err != nil {
		return err
	}

	

	return nil
}

func matched(productid int64, id *int) (int, int) {
	var count, cartid int

	err := db.QueryRow(`select product_count,cartid from cart where 
  productid=$1 and userid=$2  `,
		productid, id).Scan(&count, &cartid)
	if err != nil {

		return count, cartid
	}

	return count, cartid
}
