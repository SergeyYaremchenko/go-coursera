package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"user/user"

	"github.com/mailru/easyjson"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	users := make([]user.User, 0)
	uniqBrowsers := make(map[string]struct{})

	for scanner.Scan() {
		user := user.User{}
		err := easyjson.Unmarshal([]byte(scanner.Text()), &user)
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	builder := strings.Builder{}

	for i, user := range users {
		hasAndroid := false
		hasIe := false

		for _, browser := range user.Browsers {
			count := false
			if strings.Contains(browser, "Android") {
				hasAndroid = true
				count = true
			}

			if strings.Contains(browser, "MSIE") {
				hasIe = true
				count = true
			}

			if count {
				if _, seen := uniqBrowsers[browser]; !seen {
					uniqBrowsers[browser] = struct{}{}
				}
			}
		}

		if !(hasAndroid && hasIe) {
			continue
		}

		email := strings.ReplaceAll(user.Email, "@", " [at] ")
		builder.WriteString(fmt.Sprintf("[%d] %s <%s>\n", i, user.Name, email))
	}

	fmt.Fprintln(out, "found users:\n"+builder.String())
	fmt.Fprintln(out, "Total unique browsers", len(uniqBrowsers))
}
