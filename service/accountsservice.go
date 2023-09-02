package service

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/perisynctechnologies/pms/model"
)

var JwtKey = []byte("my_secret_key")

func RegisterUser(body model.Register) error {
	if err := body.Validate(); err != nil {
		return err
	}
	found := chekEmail(body)
	if found {
		return errors.New("email already exsits")
	}
	var userid int
	err := db.QueryRow(`INSERT INTO userdata (first_name, middle_name,last_name, email, mobile, password, created_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7) returning id`,
		body.FirstName, body.MiddleName, body.LastName, strings.ToLower(body.Email), body.MobileNumber, body.Password, time.Now()).Scan(&userid)
	if err != nil {
		return err
	}

	return nil
}
func LoginUser(body model.Login) (*model.LogUser, error) {
	if err := body.Validate(); err != nil {
		return nil, err
	}
	var user model.LogUser
	var Userid int
	err := db.QueryRow(`SELECT id,first_name,middle_name,last_name,email,created_at
	FROM userdata 
	WHERE email=$1 
	AND password=$2`, strings.ToLower(body.Email), body.Password).Scan(&Userid, &user.FirstName, &user.MiddleName, &user.LastName, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, errors.New("email and password not found")
	}

	expirationTime := time.Now().Add(15 * time.Hour)
	claims := &model.Claims{
		UserId: Userid,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		return nil, err
	}
	user.Token = tokenString
	return &user, nil

}

func chekEmail(body model.Register) bool {
	var email string
	row := db.QueryRow(`SELECT email
	FROM userdata WHERE email=$1 `,
		strings.ToLower(body.Email))

	err := row.Scan(&email)
	if err != nil {

		return false
	}
	return true
}
