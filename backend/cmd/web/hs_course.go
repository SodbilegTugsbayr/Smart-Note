package main

import (
	"net/http"
	"strconv"

	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/common/oapi"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
)

func getCourse(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	size, _ := strconv.Atoi(q.Get("size"))

	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 25
	}

	filter := new(userman.Filter)
	filter.Role = q.Get("role")
	filter.Keyword = q.Get("keyword")

	users, total, err := app.Users.GetAll(filter, page, size)
	if err != nil {
		oapi.ServerError(w, err)
		return
	}

	oapi.SendResp(w, map[string]interface{}{
		"items": users,
		"total": total,
	})
}
