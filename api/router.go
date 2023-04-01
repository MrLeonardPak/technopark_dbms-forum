package api

import (
	"github.com/MrLeonardPak/technopark_forum-dbms/middlewares"
	"github.com/fasthttp/router"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/valyala/fasthttp"
)

var DBS *pgxpool.Pool

const (
	forumSlug    = "slug"
	postSlug     = "id"
	threadSlug   = "slug_or_id"
	usernameSlug = "nickname"
)

func initForum(g *router.Group) {
	const (
		create        = "/create"
		subUrl        = "/{" + forumSlug + "}"
		details       = subUrl + "/details"
		sluggedCreate = subUrl + "/create"
		users         = subUrl + "/users"
		threads       = subUrl + "/threads"
	)
	g.POST(create, apiMiddleware(CreateForum))
	g.GET(details, apiMiddleware(GetForumDetails))
	g.POST(sluggedCreate, apiMiddleware(CreateForumThread))
	g.GET(users, apiMiddleware(GetForumUsers))
	g.GET(threads, apiMiddleware(GetForumThreads))
}

func initPost(g *router.Group) {
	const (
		postDetails = "/{" + postSlug + "}/details"
	)
	g.GET(postDetails, apiMiddleware(GetPostDetails))
	g.POST(postDetails, apiMiddleware(UpdatePostDetails))
}

func initService(g *router.Group) {
	const (
		clear  = "/clear"
		status = "/status"
	)
	g.GET(status, apiMiddleware(GetServiceStatus))
	g.POST(clear, apiMiddleware(ClearServiceData))
}

func initThread(g *router.Group) {
	const (
		subUrl  = "/{" + threadSlug + "}"
		create  = subUrl + "/create"
		details = subUrl + "/details"
		posts   = subUrl + "/posts"
		vote    = subUrl + "/vote"
	)
	g.POST(create, apiMiddleware(CreateThreadPost))
	g.GET(details, apiMiddleware(GetThreadDetails))
	g.POST(details, apiMiddleware(UpdateThreadDetails))
	g.GET(posts, apiMiddleware(GetThreadPosts))
	g.POST(vote, apiMiddleware(SetThreadVote))
}

func initUser(g *router.Group) {
	const (
		subUrl  = "/{" + usernameSlug + "}"
		create  = subUrl + "/create"
		profile = subUrl + "/profile"
	)
	g.POST(create, apiMiddleware(CreateUser))
	g.GET(profile, apiMiddleware(GetUserProfile))
	g.POST(profile, apiMiddleware(UpdateUserProfile))
}

func apiMiddleware(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return middlewares.WrapperRPS(middlewares.WrapperHeader(handler))
}

func InitRouters(g *router.Group) {
	const (
		forum   = "/forum"
		post    = "/post"
		service = "/service"
		thread  = "/thread"
		user    = "/user"
	)
	initForum(g.Group(forum))
	initPost(g.Group(post))
	initService(g.Group(service))
	initThread(g.Group(thread))
	initUser(g.Group(user))
}
