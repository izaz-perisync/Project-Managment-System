package service

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/perisynctechnologies/pms/model"
	"github.com/perisynctechnologies/pms/utils"
	gomail "gopkg.in/mail.v2"
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
	query := `select product_name,description,brand,category,price,stock from product where  product_id =$1  `
	rows, err := db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var data model.Singleproduct
	for rows.Next() {

		err := rows.Scan(&data.ProductName, &data.Description, &data.Brand, &data.Category, &data.Price, &data.Stock)
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
		}

	}
	return &data, nil
	// return nil, nil
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

func AddTocart(body model.FilterProduct, token string) error {
	id, err := VaildSign(token)
	if err != nil {
		return errors.New("token error")
	}
	var productexsits, stock, price int
	err = db.QueryRow(`select product_id,stock,price from product where product_id=$1`, body.ProductId).Scan(&productexsits, &stock, &price)
	if err != nil {
		return errors.New("product not found")
	}

	// if stock <= 0 {
	// 	return errors.New("product-out-of-stock")
	// }

	count, cartid := matched(int64(body.ProductId), id)

	// if body.Quantity > stock || count > stock {
	// 	fmt.Println(body.Quantity)
	// 	a := fmt.Sprintf("sorry! We  have any only %d units for this item", stock)
	// 	return errors.New(a)
	// }

	if count > 0 && body.Quantity == 0 {

		_, err = db.Exec(`update cart set product_count=$1 where cartid=$2`, count+1, cartid)
		if err != nil {
			return errors.New("update error")
		}
		return nil
	} else if count > 0 && body.Quantity != 0 {
		_, err = db.Exec(`update cart set product_count=$1 where cartid=$2`, count+body.Quantity, cartid)
		if err != nil {
			return errors.New("update error")
		}
		return nil
	}
	if body.Quantity != 0 {

		_, err = db.Exec(`insert into cart (productid,product_count,userid,price) values ($1,$2,$3,$4)`, productexsits, body.Quantity, id, price)
		if err != nil {
			return errors.New("insert error")
		}
		return nil
	}

	_, err = db.Exec(`insert into cart (productid,product_count,userid,price) values ($1,$2,$3,$4)`, productexsits, count+1, id, price)
	if err != nil {
		return errors.New("insert error")
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
	query := `
        SELECT c.productid, c.product_count, c.cartid, c.price,
               p.product_id, p.product_name, p.description, p.brand, p.category, p.stock
        FROM cart c
        JOIN product p ON c.productid = p.product_id
        WHERE c.userid = $1
        ORDER BY c.cartid DESC OFFSET $2 LIMIT $3
    `

	if body.Size == 0 {
		body.Size = 10
	}
	offset := 0
	if body.Page > 1 {
		offset = body.Size * (body.Page - 1)
	}

	rows, err := db.Query(query, id, offset, body.Size)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list model.CartList
	list.TotalCount = 0

	for rows.Next() {
		var body model.CartDetails
		err := rows.Scan(
			&body.ProductId, &body.ProductCount, &body.ItemId, &body.Price,
			&body.ProductId, &body.ProductName, &body.Description, &body.Brand, &body.Category, &body.Stock,
		)
		if err != nil {
			return nil, err
		}

		// Fetch assets for this product (you can batch this too if needed)

		row2, err := db.Query(`SELECT asset_id, file_path, file_type, added_at FROM product_assets WHERE productid = $1`, body.ProductId)
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

		list.TotalCount++
		list.CartData = append(list.CartData, body)
	}
	return &list, nil
}

func PlaceOrder(token string) error {
	id, err := VaildSign(token)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {

		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
		}
	}()
	var addressid int
	err = db.QueryRow(`select address_id from address_data where user_id=$1`, id).Scan(&addressid)
	if err != nil {

		return fmt.Errorf("please provide address")
	}
	fmt.Println("addressid", addressid, "userid", *id)

	var args []interface{}
	query := "SELECT cartid, productid, price, product_count FROM cart WHERE userid = $1"
	args = append(args, *id)
	rows, err := tx.Query(query, args...)
	fmt.Println(query)
	if err != nil {
		tx.Rollback()

		return err
	}
	defer rows.Close()
	var cartids []int
	var products []int
	var details []int

	var product_count []int

	for rows.Next() {
		var cartid, productid, pricedata, productcount int

		if err := rows.Scan(&cartid, &productid, &pricedata, &productcount); err != nil {
			tx.Rollback()

			return err
		}
		cartids = append(cartids, cartid)
		products = append(products, productid)
		// price = append(price, pricedata)
		product_count = append(product_count, productcount)
	}
	if len(cartids) == 0 {
		tx.Rollback()

		return errors.New("cart is empty")
	}

	if len(products) == 0 {
		tx.Rollback()

		return errors.New("few products are out-of-stock please update cart")
	}
	var stock []int
	for _, v := range products {
		var productstock int
		err = tx.QueryRow(`SELECT stock FROM product WHERE product_id = $1`, v).Scan(&productstock)
		if err != nil {
			tx.Rollback()

			return err
		}
		stock = append(stock, productstock)
	}

	for i, s := range stock {
		if s <= 0 {
			tx.Rollback()

			return fmt.Errorf("product with ID %d is out of stock", products[i])
		}
		if s < product_count[i] {
			tx.Rollback()

			return fmt.Errorf("available stock of product with ID %d is %d", products[i], s)
		}
	}

	var order_id int
	err = db.QueryRow(`insert into orders(user_id,created_at,order_status,updated_at) values ($1,$2,$3,$4) returning order_id`, id, time.Now(), utils.OrderStatus["Active"], time.Now()).Scan(&order_id)
	if err != nil {
		tx.Rollback()

		return err
	}
	var item_info []struct {
		OrderID     int
		UserID      int
		ProductId   int
		VendorId    int
		ProductName string
		Description string
		Brand       string
		Category    string
		Price       int
		Quantity    int
	}
	details = append(details, *id)
	productinfo, err := tx.Query(`select o.order_id,o.user_id,
	 p.product_id,p.product_name, 
	p.description, p.brand, p.category,p.price,p.userid,c.product_count
	  from cart c join orders o on c.userid=o.user_id 
	  JOIN product p ON p.product_id = c.productid
	  where o.user_id = $1 and o.order_id=$2`, *id, order_id)
	if err != nil {
		fmt.Println("er1")
		tx.Rollback()

		return err
	}
	defer productinfo.Close()
	for productinfo.Next() {
		var info struct {
			OrderID     int
			UserID      int
			ProductId   int
			VendorId    int
			ProductName string
			Description string
			Brand       string
			Category    string
			Price       int
			Quantity    int
		}
		err := productinfo.Scan(&info.OrderID, &info.UserID, &info.ProductId, &info.ProductName, &info.Description, &info.Brand, &info.Category, &info.Price, &info.VendorId, &info.Quantity)
		if err != nil {
			fmt.Println("er2")
			tx.Rollback()

			return err
		}

		item_info = append(item_info, info)
		fmt.Println("len", len(item_info))
	}
	for _, info := range item_info {
		fmt.Println("item info22", item_info)
		productDetails := model.ListWithoutStock{
			ProductId:   info.ProductId,
			ProductName: info.ProductName,
			Description: info.Description,
			Brand:       info.Brand,
			Category:    info.Category,
			Price:       info.Price,
		}
		productDetailsJSON, err := json.Marshal(productDetails)
		if err != nil {
			fmt.Println("er4")
			tx.Rollback()

			return err
		}
		details = append(details, info.VendorId)
		_, err = tx.Exec(`insert into orderitems (order_id,item_info,quantity,vendor_id,item_order_status) 
		values($1,$2,$3,$4,$5) `, info.OrderID, productDetailsJSON, info.Quantity, info.VendorId, utils.Active["placed"])
		if err != nil {
			fmt.Println("err", err)
			tx.Rollback()

			return err
		}

	}
	for i, pid := range products {
		fmt.Println("product_count", product_count[i])
		_, err := tx.Exec(`UPDATE product SET stock = stock - $1 WHERE product_id = $2`, product_count[i], pid)
		if err != nil {
			fmt.Println("er5")
			tx.Rollback()

			return err
		}

		_, err = tx.Exec(`DELETE FROM cart WHERE productid = $1`, pid)
		if err != nil {
			fmt.Println("er6")
			tx.Rollback()

			return err
		}
	}

	// return nil

	// fmt.Println("data", item_info)
	err = tx.Commit()
	if err != nil {
		fmt.Println("er7")
		tx.Rollback()
		return err
	}
	fmt.Println(details)
	r := Sendemail(details)
	fmt.Println(details, r)

	return nil
}

func Sendemail(details []int) error {

	for index := range details {

		m := gomail.NewMessage()

		m.SetHeader("From", "izaz@perisync.com")
		var recepient string
		fmt.Println(details[index])
		// var emailBody strings.Builder
		emailBody := `  <html>
<head>
	<style>
		table {
			border-collapse: collapse;
			width: 100%;
		}
		th, td {
			border: 1px solid #ddd;
			padding: 8px;
			text-align: left;
			
		}
		th {
			background-color: #f2f2f2;
		}
		td{
			background-colour:#7529f;
		}
	</style>
</head>
<body>
	<h2>Order Details</h2>
	<table>
		<tr>
			<th>Product Name</th>
			<th>Category</th>
			<th>Brand</th>
			<th>Price</th>
			<th>Quantity Ordered</th>
			<th>Total Price</th>
			<th>OrderStatus</th>
		</tr>`
		// emailBody.WriteString("<html><body>")
		// emailBody.WriteString("<h2>Order Details</h2>")
		// emailBody.WriteString("<table border=\"2\">")
		// emailBody.WriteString("<tr>")
		// emailBody.WriteString("<th>Product Name</th>")
		// emailBody.WriteString("<th>Category</th>")
		// emailBody.WriteString("<th>Brand</th>")
		// emailBody.WriteString("<th>Price</th>")
		// emailBody.WriteString("<th>Quantity Ordered</th>")
		// emailBody.WriteString("<th>Total Price</th>")
		// emailBody.WriteString("</tr>")
		err := db.QueryRow(`select email from userdata where id=$1`, details[index]).Scan(&recepient)
		if err != nil {
			fmt.Println("error at start", err)
			return err
		}

		m.SetHeader("To", recepient)

		m.SetHeader("Subject", "orderPlaced")

		var orderid int
		err = db.QueryRow(`select order_id from orders where user_id=$1`, details[index]).Scan(&orderid)
		if err != nil {
			fmt.Println("eq2", err)
		}
		if orderid == 0 {
			fmt.Println("orderid", orderid)
			rows, err := db.Query(`select item_info,quantity,item_order_status from orderitems where vendor_id=$1`, details[index])
			if err != nil {
				fmt.Println("errdata", err)
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var productDetailsJSON []byte
				var quantity int
				var status string
				err := rows.Scan(
					&productDetailsJSON,
					&quantity,
					&status,
				)
				if err != nil {
					fmt.Println("==", err)
					return err
				}
				var productDetails model.ListWithoutStock
				if err := json.Unmarshal(productDetailsJSON, &productDetails); err != nil {
					fmt.Println("unmarshal err", err)
					return err
				}
				fmt.Println("json", productDetails)
				fmt.Println("quantity", quantity)

				totalprice := productDetails.Price * quantity
				// m.SetBody("text/plain", "product name"+productDetails.ProductName+"productCategory:"+productDetails.Category+"Brand:"+productDetails.Brand+"price:"+strconv.Itoa(productDetails.Price)+"quantity ordered:"+strconv.Itoa(quantity)+"total price:"+strconv.Itoa(totalprice))
				// emailBody.WriteString(fmt.Sprintf("product name: %s, productCategory: %s, Brand: %s, price: %s, quantity ordered: %s, total price: %s\n",
				// 	productDetails.ProductName, productDetails.Category, productDetails.Brand, strconv.Itoa(productDetails.Price), strconv.Itoa(quantity), strconv.Itoa(totalprice)))
				// emailBody.WriteString("<tr>")
				// emailBody.WriteString("<td>" + productDetails.ProductName + "</td>")
				// emailBody.WriteString("<td>" + productDetails.Category + "</td>")
				// emailBody.WriteString("<td>" + productDetails.Brand + "</td>")
				// emailBody.WriteString("<td>" + strconv.Itoa(productDetails.Price) + "</td>")
				// emailBody.WriteString("<td>" + strconv.Itoa(quantity) + "</td>")
				// emailBody.WriteString("<td>" + strconv.Itoa(totalprice) + "</td>")
				// emailBody.WriteString("</tr>")
				emailBody += `
                <tr>
                    <td>` + productDetails.ProductName + `</td>
                    <td>` + productDetails.Category + `</td>
                    <td>` + productDetails.Brand + `</td>
                    <td>` + strconv.Itoa(productDetails.Price) + `</td>
                    <td>` + strconv.Itoa(quantity) + `</td>
                    <td>` + strconv.Itoa(totalprice) + `</td>
					<td>` + status + `</td>
                </tr>
            `
			}
		} else {
			fmt.Println("row", orderid)
			rows, err := db.Query(`select item_info,quantity,item_order_status from orderitems where order_id=$1`, orderid)
			if err != nil {
				fmt.Println("er1", err)
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var productDetailsJSON []byte
				var quantity int
				var status string
				err := rows.Scan(
					&productDetailsJSON,
					&quantity,
					&status,
				)
				if err != nil {
					fmt.Println("er2", err)
					return err
				}
				var productDetails model.ListWithoutStock
				if err := json.Unmarshal(productDetailsJSON, &productDetails); err != nil {
					fmt.Println("err145", err)
					return err
				}
				fmt.Println("====", productDetails)

				totalprice := productDetails.Price * quantity
				// m.SetBody("text/plain", "product name"+productDetails.ProductName+"productCategory:"+productDetails.Category+"Brand:"+productDetails.Brand+"price:"+strconv.Itoa(productDetails.Price)+"quantity ordered:"+strconv.Itoa(quantity)+"total price:"+strconv.Itoa(totalprice))
				// emailBody.WriteString(fmt.Sprintf("product name: %s, productCategory: %s, Brand: %s, price: %s, quantity ordered: %s, total price: %s\n",
				// 	productDetails.ProductName, productDetails.Category, productDetails.Brand, strconv.Itoa(productDetails.Price), strconv.Itoa(quantity), strconv.Itoa(totalprice)))
				// emailBody.WriteString("<tr>")
				// emailBody.WriteString("<td>" + productDetails.ProductName + "</td>")
				// emailBody.WriteString("<td>" + productDetails.Category + "</td>")
				// emailBody.WriteString("<td>" + productDetails.Brand + "</td>")
				// emailBody.WriteString("<td>" + strconv.Itoa(productDetails.Price) + "</td>")
				// emailBody.WriteString("<td>" + strconv.Itoa(quantity) + "</td>")
				// emailBody.WriteString("<td>" + strconv.Itoa(totalprice) + "</td>")
				// emailBody.WriteString("</tr>")
				emailBody += `
                <tr>
                    <td>` + productDetails.ProductName + `</td>
                    <td>` + productDetails.Category + `</td>
                    <td>` + productDetails.Brand + `</td>
                    <td>` + strconv.Itoa(productDetails.Price) + `</td>
                    <td>` + strconv.Itoa(quantity) + `</td>
                    <td>` + strconv.Itoa(totalprice) + `</td>
					<td>` + status + `</td>
                </tr>
            `

			}
		}

		// emailBody.WriteString("</table>")
		// emailBody.WriteString("</body></html>")
		// m.SetBody("text/html", emailBody.String())
		emailBody += `
                </table>
            </body>
            </html>
        `

		// Set the email body as HTML
		m.SetBody("text/html", emailBody)
		d := gomail.NewDialer("mail.perisync.com", 587, "izaz@perisync.com", "Pass@124!")
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

		// Now send E-Mail
		if err := d.DialAndSend(m); err != nil {
			fmt.Println("er1243", err)
			panic(err)
		}
	}
	return nil
}

func UpdateQuantity(token string, body model.FilterProduct) error {
	id, err := VaildSign(token)
	if err != nil {
		return err
	}
	result, err := db.Exec(`update cart set product_count=$1 where cartid=$2 and userid=$3`, body.Quantity, body.ItemId, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {

		return err
	}
	if rowsAffected == 0 {

		return fmt.Errorf("cart is empty")
	}
	return nil
}

func OrderRequest(token string, body model.FilterProduct) (*model.OrderRequest, error) {
	id, err := VaildSign(token)
	if err != nil {
		return nil, err
	}
	var list model.OrderRequest
	list.TotalPrice = 0

	baseQuery := `
	SELECT 
    r.order_id,
    r.order_item_id,
    r.quantity ,
    r.item_order_status ,
    r.item_info,
	o.created_at,
    u.first_name,
    u.middle_name,
    u.last_name,
    u.mobile,
    u.email,
	a.address_id,
	a.street_address,
   a.city,
   a.state,
   a.postal_code,
    a.country,
    a.label,
   a.latitude,
  a.longitude

FROM orderitems r
join orders o on o.order_id =r.order_id  
JOIN userdata u ON u.id = o.user_id  
JOIN address_data a on a.user_id=o.user_id
where r.vendor_id  = $1;
`
	if body.SortBy != "" {
		var sort string
		for key, val := range utils.Active {
			if body.SortBy == key {
				sort = val
			}
		}

		baseQuery += " AND r.order_status = " + (sort)
	}
	row, err := db.Query(baseQuery, id)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	for row.Next() {
		var order model.Orders
		// var orderPlaceduser int
		var productDetailsJSON []byte
		var orderstatus string
		if err := row.Scan(

			&order.OrderId,
			&order.OrderItemId,

			&order.Quantity,

			&orderstatus,
			&productDetailsJSON,
			&order.ProductData.OrderedAt,
			&order.Userdetails.FirstName,
			&order.Userdetails.MiddleName,
			&order.Userdetails.LastName,
			&order.Userdetails.MobileNumber,
			&order.Userdetails.Email,
			&order.Userdetails.AddressId,
			&order.Userdetails.Address.Street_Address,
			&order.Userdetails.Address.City,
			&order.Userdetails.Address.State,
			&order.Userdetails.Address.PostalCode,
			&order.Userdetails.Address.Country,
			&order.Userdetails.Address.Label,
			&order.Userdetails.Address.Latitude,
			&order.Userdetails.Address.Longitude); err != nil {
			return nil, err
		}
		for key, val := range utils.Active {
			if orderstatus == val {
				order.OrderStatus = key
			}
		}
		var productDetails model.ListWithoutStock
		if err := json.Unmarshal(productDetailsJSON, &productDetails); err != nil {
			return nil, err
		}
		order.ProductData.ListWithoutStock = productDetails
		row3, err := db.Query(`SELECT asset_id, file_path, file_type, added_at FROM product_assets WHERE productid = $1`, order.ProductId)
		if err != nil {
			return nil, err
		}
		defer row3.Close()
		for row3.Next() {
			var assetlist model.Assets
			err := row3.Scan(&assetlist.AssetId, &assetlist.FilePath, &assetlist.AssetType, &assetlist.Added_at)
			if err != nil {
				return nil, err
			}
			order.Assets = append(order.Assets, assetlist)
		}

		list.OrderDetails = append(list.OrderDetails, order)
		list.TotalPrice += order.Price * order.Quantity
	}
	list.TotalCount = len(list.OrderDetails)
	return &list, nil
}

func ChangeStatus(token string, orderId int64, status string) (string, error) {
	id, err := VaildSign(token)
	if err != nil {
		return "", err
	}
	var orderexsits, userexsits, productcount, productId, orderid int
	var previoustatus string
	err = db.QueryRow(`SELECT o.order_item_id, p.userid, o.item_order_status,o.quantity,(o.item_info->>'productId')::int,o.order_id
	FROM orderitems o
	JOIN product p ON (o.item_info->>'productId')::int = p.product_id
	WHERE o.order_item_id = $1 AND p.userid = $2;
	 `, orderId, id).Scan(&orderexsits, &userexsits, &previoustatus, &productcount, &productId, &orderid)
	if err != nil {
		fmt.Println("here")
		return "error in query", err
	}
	fmt.Println("itemid", orderexsits)
	value, err := conform(status)
	if err != nil {
		return "error in status", err
	}
	var details []struct {
		UserID      int
		OrderId     int
		OrderItemId int
	}
	if previoustatus == "placed" || previoustatus == "reject" {
		fmt.Println("came inside", previoustatus, value)
		if value == "accept" {

			var info struct {
				UserID      int
				OrderId     int
				OrderItemId int
			}
			_, err = db.Exec(`update orderitems set item_order_status=$1 where order_item_id=$2 `, value, orderexsits)
			if err != nil {

				return "", fmt.Errorf("error in update userorders value1")
			}
			info.OrderItemId = orderexsits
			// var ordererduser,orderid,orderitemid int
			err = db.QueryRow(`update orders set order_status=$1,updated_at=$2 where order_id=$3 returning user_id `, utils.OrderStatus["Active"], time.Now(), orderid).Scan(&info.UserID)
			if err != nil {

				return "", fmt.Errorf("error in orders val1")
			}
			info.OrderId = orderid

			details = append(details, info)
			r := StatusEmail(details)
			fmt.Println(details, r)
			return "order accepted", nil
		} else if value == "reject" {
			var info struct {
				UserID      int
				OrderId     int
				OrderItemId int
			}
			_, err = db.Exec(`update orderitems set item_order_status=$1 where order_item_id=$2`, value, orderexsits)
			if err != nil {

				return "", fmt.Errorf("error in update userorders value2")
			}
			info.OrderItemId = orderexsits

			err = db.QueryRow(`update orders set order_status=$1,updated_at=$2 where order_id=$3 returning user_id`, utils.OrderStatus["Closed"], time.Now(), orderid).Scan(&info.UserID)
			if err != nil {
				fmt.Println(err)
				return "", fmt.Errorf("error in update orders val2")
			}
			info.OrderId = orderid

			details = append(details, info)
			r := StatusEmail(details)
			fmt.Println(details, r)
			_, err := db.Exec(`update product set stock=stock+$1 where product_id =$2`, productcount, productId)
			if err != nil {
				fmt.Println("update stock err", err)

				return "", fmt.Errorf("error in update orders val2")
			}

			return "order rejected", nil

		} else {
			return "", fmt.Errorf("please check the order status")
		}

	} else if previoustatus == "accept" {
		if value == "dispatch" {

			var info struct {
				UserID      int
				OrderId     int
				OrderItemId int
			}
			_, err = db.Exec(`update orderitems set item_order_status=$1 where order_item_id=$2 `, value, orderexsits)
			if err != nil {

				return "error in update orderitems value5 ", err
			}
			info.OrderItemId = orderexsits
			fmt.Println("info order itemid", info.OrderItemId)

			err = db.QueryRow(`update orders set updated_at=$1 where order_id=$2 returning user_id  `, time.Now(), orderid).Scan(&info.UserID)
			if err != nil {
				fmt.Println("er3", err)
				return "", fmt.Errorf("error in update orders val2")
			}
			info.OrderId = orderid
			fmt.Println("info", info)

			details = append(details, info)
			fmt.Println("==", details, "-", info)
			r := StatusEmail(details)
			fmt.Println(details, r)
			return "order dispatched", nil

		} else if value == "reject" {
			var info struct {
				UserID      int
				OrderId     int
				OrderItemId int
			}
			_, err = db.Exec(`update orderitems set item_order_status=$1 where order_item_id=$2 `, value, orderexsits)
			if err != nil {

				return "", fmt.Errorf("error in update orderitems value2")
			}
			info.OrderItemId = orderexsits

			err = db.QueryRow(`update orders set order_status=$1 ,updated_at=$2 where order_id=$3 returning user_id  `, utils.OrderStatus["Closed"], time.Now(), orderid).Scan(&info.UserID)
			if err != nil {
				fmt.Println("er2", err)

				return "", fmt.Errorf("error in update orders val2")
			}
			info.OrderId = orderid

			// err = db.QueryRow(`SELECT r.user_id, o.order_item_id FROM orders r join orderitems o on o.order_id=$1 WHERE r.order_id = $1`, info.OrderId, info.OrderId).Scan(&info.UserID, &info.OrderItemId)
			// if err != nil {
			// 	fmt.Println(err)
			// 	return "", fmt.Errorf("error in fetching updated values")
			// }
			// err = db.QueryRow(`SELECT r.user_id, o.order_item_id FROM orders r join orderitems o on o.order_id=$1 WHERE r.order_id = $1`, info.OrderId).Scan(&info.UserID, &info.OrderItemId)
			// if err != nil {
			// 	fmt.Println(err)
			// 	return "", fmt.Errorf("error in fetching updated values")
			// }
			details = append(details, info)
			fmt.Println(details)
			r := StatusEmail(details)
			fmt.Println(details, r)
			_, err := db.Exec(`update product set stock=stock+$1 where product_id =$2`, productcount, productId)
			if err != nil {
				fmt.Println("er1", err)
				return "", fmt.Errorf("error in update orders val2")
			}
			return "order rejected", nil
		} else {
			return "", fmt.Errorf("please check the order status")
		}

	} else if previoustatus == "dispatch" {
		if value == "deliver" {
			var info struct {
				UserID      int
				OrderId     int
				OrderItemId int
			}
			_, err = db.Exec(`update orderitems set item_order_status=$1 where order_item_id=$2 `, value, orderexsits)
			if err != nil {
				fmt.Println("er5", err)

				return "", fmt.Errorf("error in update orders val2")
			}
			info.OrderItemId = orderexsits

			err = db.QueryRow(`update orders set order_status=$1 ,updated_at=$2 where order_id=$3 returning user_id `, utils.OrderStatus["Active"], time.Now(), orderid).Scan(&info.UserID)
			if err != nil {
				fmt.Println(err)
				return "", fmt.Errorf("error in update orders value4")
			}
			info.OrderId = orderid

			details = append(details, info)
			r := StatusEmail(details)
			fmt.Println(details, r)

			return "order delivered ", nil

		} else {
			return "", fmt.Errorf("please check the order status")
		}

	} else if previoustatus == "deliver" {
		return "", fmt.Errorf("item already delivered")
	}

	return "", nil

}

func StatusEmail(details []struct {
	UserID      int
	OrderId     int
	OrderItemId int
}) error {
	for _, info := range details {
		m := gomail.NewMessage()

		m.SetHeader("From", "izaz@perisync.com")
		var recepient string
		fmt.Println(info)
		emailBody := `  <html>
		<head>
			<style>
				table {
					border-collapse: collapse;
					width: 100%;
				}
				th, td {
					border: 1px solid #ddd;
					padding: 8px;
					text-align: left;
					
				}
				th {
					background-color: #f2f2f2;
				}
				td{
					background-colour:#7529f;
				}
			</style>
		</head>
		<body>

			<h2>Order Details</h2>
			
			`
		err := db.QueryRow(`select email from userdata where id=$1`, info.UserID).Scan(&recepient)
		if err != nil {
			fmt.Println("error at start", err)
			return err
		}
		m.SetHeader("To", recepient)
		m.SetHeader("Subject", "OrderStatus")
		rows, err := db.Query(`select item_info,quantity,item_order_status from orderitems where order_id=$1 and order_item_id=$2`, info.OrderId, info.OrderItemId)
		if err != nil {
			fmt.Println("er1", err)
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var productDetailsJSON []byte
			var quantity int
			var status string
			err := rows.Scan(
				&productDetailsJSON,
				&quantity,
				&status,
			)
			if err != nil {
				fmt.Println("er2", err)
				return err
			}
			var productDetails model.ListWithoutStock
			if err := json.Unmarshal(productDetailsJSON, &productDetails); err != nil {
				fmt.Println("err145", err)
				return err
			}
			fmt.Println("====", productDetails)

			totalprice := productDetails.Price * quantity
			emailBody += `<p><h4 style="font-size:12px ">Ordered Product </h4> ` + productDetails.ProductName + `  <h2 style="background-color: #f2f2f2">status has changed to<h2> ` + status + ` , whose brand name is ` + productDetails.Brand + `
			type of product is ` + productDetails.Category + `total amount payed for product is ` + strconv.Itoa(totalprice) + ` quantity of placed order is
			` + strconv.Itoa(quantity) + `</p>
			
               
			<table>
			<tr>
				<th>Product Name</th>
				<th>Category</th>
				<th>Brand</th>
				<th>Price</th>
				<th>Quantity Ordered</th>
				<th>Total Price</th>
				<th>OrderStatus</th>
			</tr>

                <tr>
                    <td>` + productDetails.ProductName + `</td>
                    <td>` + productDetails.Category + `</td>
                    <td>` + productDetails.Brand + `</td>
                    <td>` + strconv.Itoa(productDetails.Price) + `</td>
                    <td>` + strconv.Itoa(quantity) + `</td>
                    <td>` + strconv.Itoa(totalprice) + `</td>
					<td>` + status + `</td>
                </tr>
				
				</table>
            `

		}
		emailBody += `
            </body>
            </html>
        `
		m.SetBody("text/html", emailBody)
		d := gomail.NewDialer("mail.perisync.com", 587, "izaz@perisync.com", "Pass@124!")
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

		// Now send E-Mail
		if err := d.DialAndSend(m); err != nil {
			fmt.Println("er1243", err)
			panic(err)
		}
	}
	return nil
}

func conform(status string) (string, error) {
	for key, value := range utils.Active {
		if status == key {

			return value, nil
		}
	}
	return "", fmt.Errorf("not a valid orderstatus")
}

func OrderDetails(orderid int64) (*model.OrderData, error) {
	// Query to fetch order details and associated product details
	query := `
	SELECT
	
	r.order_item_id,
	   
	r.quantity,
   
	r.item_order_status,
	o.created_at,
	r.item_info
		FROM
			orderitems r
         join orders o on o.order_id=r.order_id		
		WHERE
			r.order_item_id = $1
    `

	rows, err := db.Query(query, orderid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var order model.OrderData
	order.TotalPrice = 0

	for rows.Next() {
		var OrderDetails model.ProductData
		var productDetailsJSON []byte
		var conform string

		err := rows.Scan(
			&OrderDetails.ProductId,
			&OrderDetails.Quantity,

			&conform,
			&OrderDetails.OrderedAt,
			&productDetailsJSON,
		)
		if err != nil {
			return nil, err
		}

		// Map the order confirmation status to its string representation
		for key, value := range utils.Active {
			if conform == value {
				OrderDetails.OrderStatus = key
			}
		}
		var productDetails model.ListWithoutStock
		if err := json.Unmarshal(productDetailsJSON, &productDetails); err != nil {
			return nil, err
		}
		OrderDetails.ListWithoutStock = productDetails

		// Calculate the total price for this product

		// Fetch product assets for this product
		rows3, err := db.Query(`
            SELECT asset_id, file_path, file_type, added_at
            FROM product_assets
            WHERE productid = $1
        `, OrderDetails.ProductId)
		if err != nil {
			return nil, err
		}
		defer rows3.Close()

		for rows3.Next() {
			var assetlist model.Assets
			err := rows3.Scan(&assetlist.AssetId, &assetlist.FilePath, &assetlist.AssetType, &assetlist.Added_at)
			if err != nil {
				return nil, err
			}

			OrderDetails.Assets = append(OrderDetails.Assets, assetlist)
		}

		order.TotalPrice += OrderDetails.Price * OrderDetails.Quantity

		order.ProductInfo = append(order.ProductInfo, OrderDetails)
	}

	order.TotalCount = len(order.ProductInfo)
	return &order, nil
}

func OrderList(token string) (*model.OrderList, error) {
	id, err := VaildSign(token)
	if err != nil {
		return nil, err
	}

	// Query to fetch order details and associated product details
	query := `
	SELECT
	   r.order_id,
	    r.order_item_id,
	   
	    r.quantity,
	   
		r.item_order_status,
		o.created_at,
	    r.item_info
			FROM
			    orderitems r join orders o on o.order_id=r.order_id
			
			WHERE
			    o.user_id = $1;
			`

	rows, err := db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list model.OrderList
	list.TotalPrice = 0

	for rows.Next() {
		var order model.OrderDetails
		var productDetailsJSON []byte
		var productstatus string
		// var status string
		err := rows.Scan(
			&order.OrderId,
			&order.OrderItemId,
			&order.Quantity,
			&productstatus,
			&order.OrderedAt,
			&productDetailsJSON,
		)
		if err != nil {
			return nil, err
		}
		for key, val := range utils.Active {
			if productstatus == val {
				order.OrderStatus = key
			}

		}
		var productDetails model.ListWithoutStock
		if err := json.Unmarshal(productDetailsJSON, &productDetails); err != nil {
			return nil, err
		}
		order.ListWithoutStock = productDetails

		rows3, err := db.Query(`
            SELECT asset_id, file_path, file_type, added_at
            FROM product_assets
            WHERE productid = $1
        `, order.ProductId)
		if err != nil {
			return nil, err
		}
		defer rows3.Close()
		for rows3.Next() {
			var assetlist model.Assets
			err := rows3.Scan(&assetlist.AssetId, &assetlist.FilePath, &assetlist.AssetType, &assetlist.Added_at)
			if err != nil {
				return nil, err
			}
			order.Assets = append(order.Assets, assetlist)
			fmt.Println(len(order.Assets))
		}

		list.OrderDetails = append(list.OrderDetails, order)
		list.TotalPrice += order.Price * order.Quantity
	}

	list.TotalCount = len(list.OrderDetails)
	return &list, nil
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
