package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"time"

	"github.com/cjp2600/structify/example/db"
)

func main() {
	connection, err := db.Open(db.Dsn(
		"localhost",
		5432,
		"test",
		"test",
		"testdb",
		"disable",
		0,
	))
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	store := db.NewBlogStorages(connection)
	userStorage := store.GetUserStorage()

	//
	//  FindMany
	//

	builder := db.FilterBuilder(
		db.UserAgeGreaterThan(20),
	) // where age > 20

	users, err := userStorage.FindMany(ctx, builder)
	if err != nil {
		if errors.Is(err, db.ErrRowNotFound) {
			fmt.Println("user not found")
			return
		}
	}

	fmt.Println(fmt.Sprintf("User: %+v", users))

	count, err := userStorage.Count(ctx, builder)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, user := range users {
		fmt.Println(fmt.Sprintf("Name: %s, Age: %d", user.Name, user.Age))
	}

	fmt.Println(fmt.Sprintf("Count: %+v", count))
}
