package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
)

type xmlUser struct {
	Id            int
	Guid          string
	IsActive      bool
	Balance       string
	Picture       string
	Age           int
	EyeColor      string
	FirstName     string
	LastName      string
	Gender        string
	Company       string
	Email         string
	Phone         string
	Address       string
	About         string
	Registered    string
	FavoriteFruit string
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
	result := make([]User, rq.Limit)

	if err != nil {
		return nil, err
	}

	u = sortUsers(u, rq.OrderField, rq.OrderBy)

	if rq.Offset > len(u)+1 || rq.Limit > len(u) {
		return u, nil
	}

	for i := rq.Offset + 1; i < rq.Offset+1+rq.Limit; i++ {
		if rq.Query != "" && (u[i].Name == rq.Query || u[i].About == rq.Query) {
			result = append(result, u[i])
		}
	}

	return result, nil
}

func sortUsers(u []User, orderFiled string, orderBy int) []User {
	if orderBy == OrderByAsIs {
		return u
	}
	sort.Slice(u, func(i, j int) bool {
		switch orderFiled {
		case "Id":
			if orderBy == OrderByAsc {
				return u[i].Id > u[j].Id
			} else {
				return u[i].Id < u[j].Id
			}
		case "Age":
			if orderBy == OrderByAsc {
				return u[i].Age > u[j].Age
			} else {
				return u[i].Age < u[j].Age
			}
		case "Name":
			fallthrough
		default:
			if orderBy == OrderByAsc {
				return u[i].Name > u[j].Name
			} else {
				return u[i].Name < u[j].Name
			}
		}
	})

	return u
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

	var xmlUsers []xmlUser
	err = xml.Unmarshal(byteValue, &xmlUsers)

	if err != nil {
		return nil, err
	}

	result := make([]User, len(xmlUsers))

	for i := 0; i < len(xmlUsers); i++ {
		result = append(result, User{
			Id:     xmlUsers[i].Id,
			Name:   xmlUsers[i].FirstName + " " + xmlUsers[i].LastName,
			Age:    xmlUsers[i].Age,
			About:  xmlUsers[i].About,
			Gender: xmlUsers[i].Gender,
		})
	}

	return result, nil
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
