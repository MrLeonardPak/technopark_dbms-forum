package api

import (
	"context"
	"errors"
	"fmt"

	"net/http"
	"strconv"
	"strings"

	"github.com/MrLeonardPak/technopark_forum-dbms/models"
	"github.com/MrLeonardPak/technopark_forum-dbms/response"
	"github.com/go-openapi/strfmt"
	"github.com/jackc/pgtype"
	"github.com/valyala/fasthttp"
)

var errNoneThread = errors.New("None thread")

func CreateThreadPost(fastCtx *fasthttp.RequestCtx) {
	threadIdOrSlug := fastCtx.UserValue(threadSlug).(string)
	ctx := context.Background()
	threadid, ok := checkThreadByIdOrSlug(ctx, threadIdOrSlug)
	if !ok {
		fmt.Println("CreateThreadPost (1)", errNoneThread)
		response.Send(http.StatusNotFound, models.Error{
			Message: errNoneThread.Error(),
		}, fastCtx)
	}
	forum, err := getThreadForumById(ctx, threadid)
	if err != nil {
		fmt.Println("CreateThreadPost (2)", err)
		response.Send(http.StatusNotFound, models.Error{
			Message: err.Error(),
		}, fastCtx)
		return
	}

	posts := new(models.Posts)
	if err = posts.UnmarshalJSON(fastCtx.Request.Body()); err != nil {
		fmt.Println("CreateThreadPost (3)", err)
	}

	if len(*posts) == 0 {
		response.Send(http.StatusCreated, posts, fastCtx)
		return
	}

	builder := strings.Builder{}

	builder.WriteString(`insert into post 
		(parent,author,message,isedited,forum,threadid,created) values`)

	args := []any{}

	parentExists := true
	for i, post := range *posts {
		if post.Parent != 0 {
			parentPost, err := getSinglePost(ctx, post.Parent)
			if err != nil {
				fmt.Println("CreateThreadPost (3):", err, post.Parent)
				response.Send(http.StatusConflict, models.Error{
					Message: err.Error() + fmt.Sprintf(" %d", post.Parent),
				}, fastCtx)
				return
			}
			if parentPost.Thread != threadid {
				fmt.Println("CreateThreadPost (4):", err)
				response.Send(http.StatusConflict, models.Error{
					Message: "Parent in another thread",
				}, fastCtx)
				return
			}
		}
		builder.WriteString(fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,now()),",
			6*i+1, 6*i+2, 6*i+3, 6*i+4, 6*i+5, 6*i+6))
		post.IsEdited = false
		post.Forum = forum
		post.Thread = threadid
		args = append(args,
			post.Parent,
			post.Author,
			post.Message,
			post.IsEdited,
			post.Forum,
			post.Thread,
		)
	}
	if !parentExists {
		fmt.Println("CreateThreadPost (5) None parent")
		response.Send(http.StatusConflict, models.Error{
			Message: "None parent",
		}, fastCtx)
		return
	}

	sqlQuery := builder.String()
	sqlQuery = strings.TrimSuffix(sqlQuery, ",")
	sqlQuery += ` returning id,parent,author,message,
		isedited,forum,threadid,created`
	rows, err := DBS.Query(ctx, sqlQuery, args...)
	if err != nil {
		fmt.Println("CreateThreadPost (6):", err)
		response.Send(http.StatusNotFound, models.Error{
			Message: err.Error(),
		}, fastCtx)
		return
	}
	defer rows.Close()
	postsResult := new(models.Posts)

	for rows.Next() {
		post := models.Post{}
		created := pgtype.Timestamp{}
		err = rows.Scan(
			&post.Id,
			&post.Parent,
			&post.Author,
			&post.Message,
			&post.IsEdited,
			&post.Forum,
			&post.Thread,
			&created,
		)
		post.Created = strfmt.DateTime(created.Time.UTC()).String()
		if err != nil {
			fmt.Println("CreateThreadPost (8):", err)
			response.Send(http.StatusInternalServerError, models.Error{
				Message: " smth wrong",
			}, fastCtx)
			return
		}
		*postsResult = append(*postsResult, post)
	}
	if len(*postsResult) == 0 {
		response.Send(http.StatusNotFound, models.Error{
			Message: "none created",
		}, fastCtx)
		return
	}
	response.Send(http.StatusCreated, postsResult, fastCtx)
	return
}

func GetThreadDetails(fastCtx *fasthttp.RequestCtx) {
	threadIdOrSlug := fastCtx.UserValue(threadSlug).(string)
	slug := ""
	id, err := strconv.Atoi(threadIdOrSlug)
	if err != nil {
		id = 0
		slug = threadIdOrSlug
	}
	ctx := context.Background()
	thread := models.ThreadModel{}
	if err = DBS.QueryRow(ctx, `
		select id, title, author, forum, message, slug, votes, created
			from thread 
		where id = $1 or lower(slug) = lower($2)`, id, slug).
		Scan(
			&thread.Id,
			&thread.Title,
			&thread.Author,
			&thread.Forum,
			&thread.Message,
			&thread.Slug,
			&thread.Votes,
			&thread.Created,
		); err != nil {
		fmt.Println("GetThreadDetails:", err, slug, id)
		response.Send(http.StatusNotFound, models.Error{
			Message: err.Error(),
		}, fastCtx)
		return
	}

	response.Send(http.StatusOK, models.Thread{
		Id:      thread.Id,
		Title:   thread.Title,
		Author:  thread.Author,
		Forum:   thread.Forum,
		Message: thread.Message,
		Slug:    thread.Slug.String,
		Votes:   thread.Votes,
		Created: strfmt.DateTime(thread.Created.UTC()).String(),
	}, fastCtx)
}
func UpdateThreadDetails(fastCtx *fasthttp.RequestCtx) {
	ctx := context.Background()
	slug := fastCtx.UserValue(threadSlug).(string)
	threadid, ok := checkThreadByIdOrSlug(ctx, slug)
	if !ok {
		fmt.Println("UpdateThreadDetails (1)", errNoneThread)
		response.Send(http.StatusNotFound, models.Error{
			Message: errNoneThread.Error(),
		}, fastCtx)
		return
	}
	threadUpdate := models.ThreadUpdate{}
	if err := threadUpdate.UnmarshalJSON(fastCtx.Request.Body()); err != nil {
		fmt.Println("UpdateThreadDetails (2)", err)
		return
	}
	if threadUpdate.Message == "" && threadUpdate.Title == "" {
		thread, err := getThreadById(ctx, threadid)
		if err != nil {
			response.Send(http.StatusNotFound, models.Error{
				Message: errNoneThread.Error(),
			}, fastCtx)
			return
		}
		response.Send(http.StatusOK, models.Thread{
			Id:      thread.Id,
			Title:   thread.Title,
			Author:  thread.Author,
			Forum:   thread.Forum,
			Message: thread.Message,
			Slug:    thread.Slug.String,
			Created: strfmt.DateTime(thread.Created.UTC()).String(),
		}, fastCtx)
		return
	}
	builder := strings.Builder{}
	builder.WriteString("update thread set")
	if threadUpdate.Title != "" {
		builder.WriteString(fmt.Sprintf(` title = '%s'`, threadUpdate.Title))
	}
	if threadUpdate.Message != "" {
		if threadUpdate.Title != "" {
			builder.WriteString(",")
		}
		builder.WriteString(fmt.Sprintf(` message = '%s'`, threadUpdate.Message))

	}
	builder.WriteString(" where id = $1 returning id,title,author,forum,message,slug,created")
	thread := models.ThreadModel{}
	if err := DBS.QueryRow(ctx, builder.String(), threadid).Scan(
		&thread.Id,
		&thread.Title,
		&thread.Author,
		&thread.Forum,
		&thread.Message,
		&thread.Slug,
		&thread.Created,
	); err != nil {
		fmt.Println("GetThreadDetails (3) none thread", err)
		response.Send(http.StatusNotFound, models.Error{
			Message: "None thread",
		}, fastCtx)
		return
	}
	response.Send(http.StatusOK, models.Thread{
		Id:      thread.Id,
		Title:   thread.Title,
		Author:  thread.Author,
		Forum:   thread.Forum,
		Message: thread.Message,
		Slug:    thread.Slug.String,
		Created: strfmt.DateTime(thread.Created.UTC()).String(),
	}, fastCtx)
}

func SetThreadVote(fastCtx *fasthttp.RequestCtx) {
	ctx := context.Background()
	slug := fastCtx.UserValue(threadSlug).(string)
	threadid, ok := checkThreadByIdOrSlug(ctx, slug)
	if !ok {
		fmt.Println("SetThreadVote (1)", errNoneThread, slug)
		response.Send(http.StatusNotFound, models.Error{
			Message: errNoneThread.Error(),
		}, fastCtx)
		return
	}
	vote := models.Vote{}
	if err := vote.UnmarshalJSON(fastCtx.Request.Body()); err != nil {
		fmt.Println("SetThreadVote (2)", err)
	}
	if _, err := DBS.Exec(ctx, `
		insert into vote (threadid, nickname, voice)
			values ($1,$2,$3)
		on conflict on constraint unique_voice do update 
		set voice = excluded.voice`, threadid, vote.Nickname, vote.Voice); err != nil {
		fmt.Println("SetThreadVote (3)", err)
		response.Send(http.StatusNotFound, models.Error{
			Message: "None thread",
		}, fastCtx)
		return
	}
	thread, err := getThreadById(ctx, threadid)
	if err != nil {
		fmt.Println("SetThreadVote (3)", err)
		response.Send(http.StatusNotFound, models.Error{
			Message: "None thread",
		}, fastCtx)
		return
	}
	response.Send(http.StatusOK, models.Thread{
		Id:      thread.Id,
		Title:   thread.Title,
		Author:  thread.Author,
		Forum:   thread.Forum,
		Message: thread.Message,
		Slug:    thread.Slug.String,
		Created: strfmt.DateTime(thread.Created.UTC()).String(),
		Votes:   thread.Votes,
	}, fastCtx)
}

func GetThreadPosts(fastCtx *fasthttp.RequestCtx) {
	ctx := context.Background()
	slug := fastCtx.UserValue(threadSlug).(string)
	threadid, ok := checkThreadByIdOrSlug(ctx, slug)
	if !ok {
		fmt.Println("GetThreadPosts (1)", errNoneThread)
		response.Send(http.StatusNotFound, models.Error{
			Message: errNoneThread.Error(),
		}, fastCtx)
		return
	}

	limitString := string(fastCtx.QueryArgs().Peek("limit"))
	sinceString := string(fastCtx.QueryArgs().Peek("since"))
	descString := string(fastCtx.QueryArgs().Peek("desc"))
	limit := 0
	since := 0
	desc := false
	var err error
	if limitString != "" {
		if limit, err = strconv.Atoi(limitString); err != nil {
			fmt.Println("GetThreadPosts (2)", err)
			response.Send(http.StatusInternalServerError, models.Error{
				Message: " smth wrong",
			}, fastCtx)
			return
		}
	}
	if sinceString != "" {
		if since, err = strconv.Atoi(sinceString); err != nil {
			fmt.Println("GetThreadPosts (3)", err)
			response.Send(http.StatusInternalServerError, models.Error{
				Message: " smth wrong",
			}, fastCtx)
			return
		}
	}
	if descString != "" {
		if desc, err = strconv.ParseBool(descString); err != nil {
			response.Send(http.StatusInternalServerError, models.Error{
				Message: " smth wrong",
			}, fastCtx)
			return
		}
	}
	sort := string(fastCtx.QueryArgs().Peek("sort"))

	if sort == "" {
		sort = "flat"
	}
	posts := new(models.Posts)
	switch sort {
	case "flat":
		*posts, err = getPostsFlat(ctx, threadid, limit, since, desc)
	case "tree":
		*posts, err = getPostsTree(ctx, threadid, limit, since, desc)
	case "parent_tree":
		*posts, err = getPostsParentTree(ctx, threadid, limit, since, desc)
	}
	if err != nil {
		fmt.Println("GetThreadPosts (4)", err)
		response.Send(http.StatusInternalServerError, models.Error{
			Message: " smth wrong",
		}, fastCtx)
		return
	}

	response.Send(http.StatusOK, posts, fastCtx)
	return
}

func checkThreadByIdOrSlug(ctx context.Context, threadIdOrSlug string) (int, bool) {
	slug := ""
	id, err := strconv.Atoi(threadIdOrSlug)
	if err != nil {
		id = 0
		slug = threadIdOrSlug
	}
	if err = DBS.QueryRow(ctx, `select id from thread 
		where id = $1 or lower(slug) = lower($2)`, id, slug).Scan(&id); err != nil {
		fmt.Println("checkThreadByIdOrSlug none thread: ", slug, id)
		return 0, false
	}
	return id, true
}

func getThreadById(ctx context.Context, threadId int) (models.ThreadModel, error) {
	thread := models.ThreadModel{}
	if err := DBS.QueryRow(ctx, `
		select id, title, author, forum, message, slug, votes, created
			from thread 
		where id = $1 `, threadId).
		Scan(
			&thread.Id,
			&thread.Title,
			&thread.Author,
			&thread.Forum,
			&thread.Message,
			&thread.Slug,
			&thread.Votes,
			&thread.Created,
		); err != nil {
		fmt.Println("SetThreadVote (1) none thread", err)
		return models.ThreadModel{}, errors.New("None thread")
	}
	return thread, nil
}

func getThreadForumById(ctx context.Context, threadId int) (string, error) {
	forum := ""
	if err := DBS.QueryRow(ctx, `
		select forum
			from thread 
		where id = $1 `, threadId).
		Scan(
			&forum,
		); err != nil {
		fmt.Println("SetThreadVote (1) none thread", err)
		return forum, errors.New("None thread")
	}
	return forum, nil
}
