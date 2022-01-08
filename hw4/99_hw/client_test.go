package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

type searchParams struct {
	Limit      int
	Offset     int
	Query      string
	OrderField string
	OrderBy    string
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	limit := r.FormValue("limit")
	offset := r.FormValue("offset")
	query := r.FormValue("query")
	orderField := r.FormValue("order_field")
	orderBy := r.FormValue("order_by")

	if err := validateParams(orderField); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rq := parseParams(limit, offset, query, orderField, orderBy)

	users, err := findUsers(rq)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	j, err := json.Marshal(users)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(j)
}

func findUsers(rq SearchRequest) ([]User, error) {
	u, err := readUsers("dataset.xml")

	if err != nil {
		return nil, err
	}

}

func readUsers(path string) ([]User, error) {
	xmlFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer xmlFile.Close()

	byteValue, err := ioutil.ReadAll(xmlFile)

	if err != nil {
		return nil, err
	}

	var users []User
	err = xml.Unmarshal(byteValue, &users)

	if err != nil {
		return nil, err
	}

	return users, nil
}

func parseParams(limit, offset, query, orderField, orderBy string) SearchRequest {
	l, err := strconv.ParseInt(limit, 10, 0)

	if err != nil || l <= 0 {
		l = 25
	}

	o, err := strconv.ParseInt(offset, 10, 0)

	if err != nil || o <= 0 {
		o = 0
	}

	ord, err := strconv.ParseInt(orderBy, 10, 0)

	if err != nil || o <= 0 {
		ord = OrderByAsIs
	}

	if orderField == "" {
		orderField = "Name"
	}

	return SearchRequest{
		Limit:      int(l),
		Offset:     int(o),
		Query:      query,
		OrderField: orderField,
		OrderBy:    int(ord),
	}
}

func validateParams(orderField string) error {
	switch orderField {
	case "Id":
		fallthrough
	case "Name":
		fallthrough
	case "":
		fallthrough
	case "Age":
		break
	default:
		return fmt.Errorf("order_field must be one of: \"Id\", \"Name\", \"Age\" or empty")
	}

	return nil
}
