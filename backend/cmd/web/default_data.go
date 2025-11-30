package main

import (
	"github.com/peakdot/go-nuxt-example/backend/cmd/web/app"
	"github.com/peakdot/go-nuxt-example/backend/pkg/entities"
	"github.com/peakdot/go-nuxt-example/backend/pkg/userman"
)

func addDefaultRecordsIfNotExist() {
	count, err := app.Users.Count(nil)
	panicOnError(err)

	if count == 0 {
		_, err := app.Users.Save(&userman.User{
			Model: entities.Model{ID: 1},
			Role:  userman.ROLE_ADMIN,
			Name:  "Orgio",
			Email: "peakdot1@gmail.com",
		})
		panicOnError(err)
	}
}
