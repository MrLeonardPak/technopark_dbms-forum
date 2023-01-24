package api

import (
	"github.com/fasthttp/router"
	"github.com/jackc/pgx/v4/pgxpool"
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
	g.POST(create, CreateForum)
	g.GET(details, GetForumDetails)
	g.POST(sluggedCreate, CreateForumThread)
	g.GET(users, GetForumUsers)
	g.GET(threads, GetForumThreads)
}

func initPost(g *router.Group) {
	const (
		postDetails = "/{" + postSlug + "}/details"
	)
	g.GET(postDetails, GetPostDetails)
	g.POST(postDetails, UpdatePostDetails)
}

func initService(g *router.Group) {
	const (
		clear  = "/clear"
		status = "/status"
	)
	g.GET(status, GetServiceStatus)
	g.POST(clear, ClearServiceData)
}

func initThread(g *router.Group) {
	const (
		subUrl  = "/{" + threadSlug + "}"
		create  = subUrl + "/create"
		details = subUrl + "/details"
		posts   = subUrl + "/posts"
		vote    = subUrl + "/vote"
	)
	g.POST(create, CreateThreadPost)
	g.GET(details, GetThreadDetails)
	g.POST(details, UpdateThreadDetails)
	g.GET(posts, GetThreadPosts)
	g.POST(vote, SetThreadVote)
}

func initUser(g *router.Group) {
	const (
		subUrl  = "/{" + usernameSlug + "}"
		create  = subUrl + "/create"
		profile = subUrl + "/profile"
	)
	g.POST(create, CreateUser)
	g.GET(profile, GetUserProfile)
	g.POST(profile, UpdateUserProfile)
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
