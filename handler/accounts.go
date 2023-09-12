package handler

import (
	"encoding/json"
	"net/http"

	"github.com/perisynctechnologies/pms/model"
	"github.com/perisynctechnologies/pms/service"
)

func writeJson(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"msg": v,
	})
}

func HandlerRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var body model.Register
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "malformed request")
		return
	}
	err = service.RegisterUser(body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJson(w, http.StatusCreated, "Registration Succesfull")
}
func HandlerLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var body model.Login
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, "malformed request")
		return
	}
	userdata, err := service.LoginUser(body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	if userdata != nil {
		json.NewEncoder(w).Encode(userdata)
		return
	}
	writeJson(w, http.StatusNoContent, "no rows found")

}

func HandlerUserAddress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	tokenstring := getTokenStringFromRequest(r)
	var body model.Address
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	err = service.AddAddress(tokenstring, body)
	if err != nil {
		writeJson(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJson(w, http.StatusCreated, "Address Added")

}
