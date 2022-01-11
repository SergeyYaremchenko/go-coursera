package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"
)

type XmlRoot struct {
	XmlName xml.Name  `xml:"root"`
	Users   []XmlUser `xml:"row"`
}

type XmlUser struct {
	XmlName       xml.Name `xml:"row"`
	Id            int      `xml:"id"`
	Guid          string   `xml:"guid"`
	IsActive      bool     `xml:"isActive"`
	Balance       string   `xml:"balance"`
	Picture       string   `xml:"picture"`
	Age           int      `xml:"age"`
	EyeColor      string   `xml:"eyeColor"`
	FirstName     string   `xml:"first_name"`
	LastName      string   `xml:"last_name"`
	Gender        string   `xml:"gender"`
	Company       string   `xml:"company"`
	Email         string   `xml:"email"`
	Phone         string   `xml:"phone"`
	Address       string   `xml:"address"`
	About         string   `xml:"about"`
	Registered    string   `xml:"registered"`
	FavoriteFruit string   `xml:"favoriteFruit"`
}

type UserApi struct {
	UserApiURL string
}

type UserTestCase struct {
	Request     SearchRequest
	Result      SearchResponse
	Token       string
	Error       error
	IsError     bool
	CheckResult func(tk UserTestCase) bool
}

func TestFindUsers(t *testing.T) {
	cases := []UserTestCase{
		{
			Request: SearchRequest{
				Query:  "",
				Limit:  10,
				Offset: 0,
			},
			Token:   "token",
			IsError: false,
			CheckResult: func(tk UserTestCase) bool {
				return len(tk.Result.Users) == tk.Request.Limit
			},
		},
		{
			Request: SearchRequest{
				Query:  "Boyd Wolf",
				Limit:  10,
				Offset: 0,
			},
			IsError: false,
			Token:   "token",
			CheckResult: func(tk UserTestCase) bool {
				return len(tk.Result.Users) == 1 && tk.Result.Users[0].Name == "Boyd Wolf"
			},
		},
		{
			Request: SearchRequest{
				Query:  "",
				Limit:  -1,
				Offset: 0,
			},
			Token:   "token",
			IsError: true,
			CheckResult: func(tk UserTestCase) bool {
				return tk.Error.Error() == "limit must be > 0"
			},
		},
		{
			Request: SearchRequest{
				Query:  "",
				Limit:  26,
				Offset: 0,
			},
			Token:   "token",
			IsError: false,
			CheckResult: func(tk UserTestCase) bool {
				return len(tk.Result.Users) == 25
			},
		},
		{
			Request: SearchRequest{
				Query:  "",
				Limit:  0,
				Offset: -1,
			},
			Token:   "token",
			IsError: true,
			CheckResult: func(tk UserTestCase) bool {
				return tk.Error.Error() == "offset must be > 0"
			},
		},
		{
			Request: SearchRequest{
				Query:  "",
				Limit:  0,
				Offset: 0,
			},
			IsError: true,
			CheckResult: func(tk UserTestCase) bool {
				return tk.Error.Error() == "Bad AccessToken"
			},
		},
		{
			Request: SearchRequest{
				Query:      "",
				Limit:      0,
				Offset:     0,
				OrderBy:    0,
				OrderField: "BAD",
			},
			IsError: true,
			Token:   "token",
			CheckResult: func(tk UserTestCase) bool {
				return tk.Error.Error() == "OrderFeld BAD invalid"
			},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	for caseNum, item := range cases {
		u := &SearchClient{
			URL:         ts.URL,
			AccessToken: item.Token,
		}
		res, err := u.FindUsers(item.Request)

		item.Error = err
		if res != nil {
			item.Result = *res
		}

		if err != nil && !item.IsError {
			t.Errorf("Unexpected error %s", err)
		}

		if err == nil && item.IsError {
			t.Errorf("Expected error, got nil: %+v", item)
		}

		if !item.CheckResult(item) {
			t.Errorf("[%d] wrong result %+v", caseNum, item)
		}
	}

	ts.Close()

	testTimeout(t)
	testInternalError(t)
	testError(t)
	testBadRequest(t)
	testBrokenJsonWithBadRequest(t)
	testBrokenJsonWithOkRequest(t)
}

func testBrokenJsonWithBadRequest(t *testing.T) {
	tsBadJsonRequest := httptest.NewServer(http.HandlerFunc(ServeBrokenJson))
	u := &SearchClient{
		URL:         tsBadJsonRequest.URL,
		AccessToken: "TOKEN",
	}
	_, err := u.FindUsers(SearchRequest{
		Limit:      10,
		Offset:     0,
		Query:      "",
		OrderField: "",
		OrderBy:    0,
	})

	if err == nil {
		t.Errorf("Expected bad json request, got nil")
	}
}

func testBrokenJsonWithOkRequest(t *testing.T) {
	tsBadJsonRequest := httptest.NewServer(http.HandlerFunc(ServeBrokenJsonOkResult))
	u := &SearchClient{
		URL:         tsBadJsonRequest.URL,
		AccessToken: "TOKEN",
	}
	_, err := u.FindUsers(SearchRequest{
		Limit:      10,
		Offset:     0,
		Query:      "",
		OrderField: "",
		OrderBy:    0,
	})

	if err == nil {
		t.Errorf("Expected bad json request, got nil")
	}
}

func testBadRequest(t *testing.T) {
	tsBadRequest := httptest.NewServer(http.HandlerFunc(ServeBadRequest))
	u := &SearchClient{
		URL:         tsBadRequest.URL,
		AccessToken: "TOKEN",
	}
	_, err := u.FindUsers(SearchRequest{
		Limit:      10,
		Offset:     0,
		Query:      "",
		OrderField: "",
		OrderBy:    0,
	})

	if err == nil {
		t.Errorf("Expected bad request, got nil")
	}

	if err != nil && err.Error() != "unknown bad request error: OUCH" {
		t.Errorf("Expected bad request, got %s", err)
	}
}

func testInternalError(t *testing.T) {
	tsInternalError := httptest.NewServer(http.HandlerFunc(ServeInternalError))
	u := &SearchClient{
		URL:         tsInternalError.URL,
		AccessToken: "TOKEN",
	}
	_, err := u.FindUsers(SearchRequest{
		Limit:      10,
		Offset:     0,
		Query:      "",
		OrderField: "",
		OrderBy:    0,
	})

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func testError(t *testing.T) {
	tsError := httptest.NewServer(http.HandlerFunc(ServeError))
	u := &SearchClient{
		URL:         tsError.URL,
		AccessToken: "TOKEN",
	}
	_, err := u.FindUsers(SearchRequest{
		Limit:      10,
		Offset:     0,
		Query:      "",
		OrderField: "",
		OrderBy:    0,
	})

	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func testTimeout(t *testing.T) {
	tsTimeoutError := httptest.NewServer(http.HandlerFunc(ServeTimeout))
	u := &SearchClient{
		URL:         tsTimeoutError.URL,
		AccessToken: "TOKEN",
	}
	_, err := u.FindUsers(SearchRequest{
		Limit:      10,
		Offset:     0,
		Query:      "",
		OrderField: "",
		OrderBy:    0,
	})

	if err == nil {
		t.Errorf("Expected timeout error, got nil")
	}
}

func ServeTimeout(w http.ResponseWriter, r *http.Request) {
	time.Sleep(5 * time.Second)
	w.WriteHeader(http.StatusOK)
}

func ServeError(w http.ResponseWriter, r *http.Request) {
	panic("ERROR")
}

func ServeInternalError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func ServeBrokenJson(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("_________brooooooken"))
}

func ServeBrokenJsonOkResult(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("_________brooooooken"))
}

func ServeBadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	resp := SearchErrorResponse{
		Error: "OUCH",
	}
	j, _ := json.Marshal(resp)
	w.Write(j)
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	limit := r.FormValue("limit")
	offset := r.FormValue("offset")
	query := r.FormValue("query")
	orderField := r.FormValue("order_field")
	orderBy := r.FormValue("order_by")

	token := r.Header.Get("AccessToken")

	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rq := parseParams(limit, offset, query, orderField, orderBy)

	if err := validateParams(rq.OrderField, rq.OrderBy); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errorBody := SearchErrorResponse{
			Error: err.Error(),
		}
		j, _ := json.Marshal(errorBody)

		w.Write(j)
		return
	}

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

	u = sortUsers(u, rq.OrderField, rq.OrderBy)

	if rq.Offset > len(u)+1 {
		return make([]User, 0), nil
	}

	u = filterUsers(u, rq.Query)
	u = trimUsers(u, rq.Offset, rq.Limit)

	return u, nil
}

func trimUsers(u []User, offset, limit int) []User {
	start := offset
	finish := start + limit

	if finish > len(u) {
		finish = len(u)
	}

	result := u[start:finish]
	return result
}

func filterUsers(u []User, query string) []User {
	if query == "" {
		return u
	}

	result := make([]User, 0)
	for i := 0; i < len(u); i++ {
		if query != "" && (u[i].Name == query || u[i].About == query) {
			result = append(result, u[i])
		}
	}
	return result
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

	var xmlRoot XmlRoot
	err = xml.Unmarshal(byteValue, &xmlRoot)

	if err != nil {
		return nil, err
	}

	xmlUsers := xmlRoot.Users

	result := make([]User, len(xmlUsers))

	for i := 0; i < len(xmlUsers); i++ {
		result[i].Id = xmlUsers[i].Id
		result[i].Name = xmlUsers[i].FirstName + " " + xmlUsers[i].LastName
		result[i].Age = xmlUsers[i].Age
		result[i].About = xmlUsers[i].About
		result[i].Gender = xmlUsers[i].Gender
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

	if err != nil {
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

func validateParams(orderField string, orderBy int) error {
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
		return fmt.Errorf("ErrorBadOrderField")
	}

	switch orderBy {
	case OrderByAsIs:
		fallthrough
	case OrderByAsc:
		fallthrough
	case OrderByDesc:
		break
	default:
		return fmt.Errorf(ErrorBadOrderField)
	}

	return nil
}
