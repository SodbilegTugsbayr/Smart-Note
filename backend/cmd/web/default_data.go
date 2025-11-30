package main

import (
	"github.com/SodbilegTugsbayr/Smart-Note/backend/cmd/web/app"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/entities"
	"github.com/SodbilegTugsbayr/Smart-Note/backend/pkg/userman"
)

func addDefaultRecordsIfNotExist() {
	count, err := app.Users.Count(nil)
	panicOnError(err)

	if count == 0 {
		_, err := app.Users.Save(&userman.User{
			Model: entities.Model{ID: 1},
			Role:  userman.ROLE_ADMIN,
			Name:  "Orgio",
			Email: "sodbileg.tu@gmail.com",
		})
		panicOnError(err)
	}
}
