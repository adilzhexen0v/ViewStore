package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	router.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fileServer))

	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf)

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/myprofile", dynamic.ThenFunc(app.myProfile))
	router.Handler(http.MethodGet, "/recommendations", dynamic.ThenFunc(app.recommendationsUsers))
	router.Handler(http.MethodGet, "/user/view/:id", dynamic.ThenFunc(app.userView))
	router.Handler(http.MethodGet, "/user/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodPost, "/user/signup", dynamic.ThenFunc(app.userSignupPost))
	router.Handler(http.MethodGet, "/user/login", dynamic.ThenFunc(app.userLogin))
	router.Handler(http.MethodPost, "/user/login", dynamic.ThenFunc(app.userLoginPost))

	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodGet, "/post/create", protected.ThenFunc(app.postCreate))
	router.Handler(http.MethodPost, "/post/newpost", protected.ThenFunc(app.postCreatePost))
	router.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.userLogoutPost))
	router.Handler(http.MethodPost, "/user/update", protected.ThenFunc(app.updateProfileData))
	router.Handler(http.MethodPost, "/user/subscribe", protected.ThenFunc(app.subscribe))
	router.Handler(http.MethodPost, "/user/unsubscribe", protected.ThenFunc(app.unsubscribe))
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	return standard.Then(router)
}
