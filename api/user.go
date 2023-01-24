package api

import (
	"context"
	"errors"
	"fmt"

	"net/http"

	"github.com/MrLeonardPak/technopark_forum-dbms/models"
	"github.com/MrLeonardPak/technopark_forum-dbms/response"
	"github.com/jackc/pgx/v4"
	"github.com/valyala/fasthttp"
)

func CreateUser(fastCtx *fasthttp.RequestCtx) {
	ctx := context.Background()
	nickname, err := getSlugUsername(fastCtx)
	if err != nil {
		fmt.Println("CreateUser (1):", err)
		return
	}
	user := new(models.User)
	if err := user.UnmarshalJSON(fastCtx.Request.Body()); err != nil {
		fmt.Println("CreateUser (2):", err)
		return
	}
	user.Nickname = nickname
	_, err = DBS.Exec(ctx, `
		insert into 
		    actor (nickname,fullname,about,email) 
		values ($1,$2,$3,$4)
		`,
		user.Nickname, user.Fullname, user.About, user.Email)
	if err == nil {
		response.Send(http.StatusCreated, user, fastCtx)
		return
	}
	rows, err := DBS.Query(ctx, `
		select nickname,fullname,about,email 
		from actor 
		where  lower(nickname) = lower($1) or lower(email) = lower($2)
		`, user.Nickname, user.Email)
	defer rows.Close()
	users := new(models.Users)
	for rows.Next() {
		rowUser := models.User{}
		if err = rows.Scan(
			&rowUser.Nickname,
			&rowUser.Fullname,
			&rowUser.About,
			&rowUser.Email,
		); err != nil {
			response.Send(http.StatusInternalServerError, models.Error{Message: "Cannot get user" + err.Error()}, fastCtx)
			return
		}
		*users = append(*users, rowUser)
	}
	response.Send(http.StatusConflict, users, fastCtx)
}

func GetUserProfile(fastCtx *fasthttp.RequestCtx) {
	nickname, err := getSlugUsername(fastCtx)
	ctx := context.Background()
	user, err := getUserByNickname(ctx, nickname)
	if err != nil {
		response.Send(http.StatusNotFound, models.Error{Message: "Not found"}, fastCtx)
		return
	}
	response.Send(http.StatusOK, user, fastCtx)
}

func getUserByNickname(ctx context.Context, nickname string) (models.User, error) {
	user := models.User{}
	err := DBS.QueryRow(ctx, ` 
			select nickname,fullname,about,email 
				from actor
			where  lower(nickname) = lower($1)`, nickname).
		Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)

	return user, err
}

func UpdateUserProfile(fastCtx *fasthttp.RequestCtx) {
	nickname, err := getSlugUsername(fastCtx)
	if err != nil {
		response.Send(http.StatusInternalServerError, models.Error{
			Message: " smth wrong",
		}, fastCtx)
		return
	}
	user := new(models.User)
	if err := user.UnmarshalJSON(fastCtx.Request.Body()); err != nil {
		fmt.Println("UpdateUserProfile:", err)
		return
	}
	user.Nickname = nickname
	userModel := models.User{}
	ctx := context.Background()
	if err = DBS.QueryRow(ctx,
		`select nickname,fullname,about,email from actor 
                where lower(nickname) = lower($1)
		`, user.Nickname).Scan(
		&userModel.Nickname,
		&userModel.Fullname,
		&userModel.About,
		&userModel.Email,
	); err == pgx.ErrNoRows {
		response.Send(http.StatusNotFound, models.Error{Message: "none such user"}, fastCtx)
		return
	}
	if user.About == "" {
		user.About = userModel.About
	}
	if user.Email == "" {
		user.Email = userModel.Email
	}
	if user.Fullname == "" {
		user.Fullname = userModel.Fullname
	}
	if err = DBS.QueryRow(ctx, `
		update actor 
		set fullname = $2,
		    about = $3,
		    email = $4
		where lower(nickname) = lower($1)
		returning nickname,fullname,about,email
		`,
		user.Nickname, user.Fullname, user.About, user.Email).
		Scan(
			&user.Nickname,
			&user.Fullname,
			&user.About,
			&user.Email,
		); err != nil {

		response.Send(http.StatusConflict, models.Error{Message: "new params don't suit"}, fastCtx)
		return
	}
	response.Send(200, user, fastCtx)
	return
}

func getSlugUsername(fastCtx *fasthttp.RequestCtx) (string, error) {
	username := fastCtx.UserValue(usernameSlug).(string)
	if username == "" {
		return username, errors.New("None user")
	}
	return username, nil
}
