package main

import (
	"errors"
	"finalexam.project/internal/models"
	"finalexam.project/internal/validator"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

type userSignupForm struct {
	Name                string `form:"name"`
	Surname             string `form:"surname"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type postCreateForm struct {
	Content             string `form:"content"`
	validator.Validator `form:"-"`
}

type subForm struct {
	Id                  int `form:"id"`
	validator.Validator `form:"-"`
}

type updateChangeDataForm struct {
	Name                string `form:"name"`
	Surname             string `form:"surname"`
	Email               string `form:"email"`
	validator.Validator `form:"-"`
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if app.getUserId(r) == 0 {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
	}
	data := app.newTemplateData(r)
	users, err := app.users.GetAllSubscriptions(app.getUserId(r))
	if err != nil {
		log.Fatal(err)
	}
	posts, err := app.posts.GetAllPosts(app.getUserId(r))
	if err != nil {
		log.Fatal(err)
	}
	data.Posts = posts
	data.Users = users
	app.render(w, http.StatusOK, "index.html", data)
}

func (app *application) postCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "createpost.html", data)
}

func (app *application) postCreatePost(w http.ResponseWriter, r *http.Request) {
	var form postCreateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.CheckField(validator.NotBlank(form.Content), "name", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "createpost.html", data)
		return
	}
	authorId := app.getUserId(r)
	postId, err := app.posts.Insert(form.Content, authorId)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = r.ParseMultipartForm(100000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	files := r.MultipartForm.File["images"]
	for i, fileHandler := range files {
		file, err := files[i].Open()
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		dst, _ := os.Create("./ui/static/media/" + fileHandler.Filename)
		if err != nil {
			log.Fatal(err)
		}
		defer dst.Close()

		io.Copy(dst, file)

		err = app.images.Insert(postId, fileHandler.Filename, authorId)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) myProfile(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	users, err := app.users.GetAllSubscriptions(app.getUserId(r))
	if err != nil {
		log.Fatal(err)
	}
	subs, err := app.users.GetAllSubscribers(app.getUserId(r))
	if err != nil {
		log.Fatal(err)
	}
	posts, err := app.posts.GetAllMyPosts(app.getUserId(r))
	if err != nil {
		errorLog.Fatal(err)
	}
	data.Posts = posts
	data.Users = users
	data.Subs = subs
	app.render(w, http.StatusOK, "myprofile.html", data)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, http.StatusOK, "signup.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Surname), "surname", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	err = app.users.Insert(form.Name, form.Surname, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, http.StatusOK, "login.html", data)

}
func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Log out")
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) updateProfileData(w http.ResponseWriter, r *http.Request) {
	var form updateChangeDataForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Surname), "surname", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "myprofile.html", data)
		return
	}

	err = r.ParseMultipartForm(100000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	file, handler, err := r.FormFile("picture")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	f, err := handler.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	dst, _ := os.Create("./ui/static/media/" + handler.Filename)
	if err != nil {
		log.Fatal(err)
	}
	defer dst.Close()

	io.Copy(dst, f)

	err = app.users.Update(form.Name, form.Surname, form.Email, handler.Filename, app.getUserId(r))
	if err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/myprofile", http.StatusSeeOther)
}

func (app *application) recommendationsUsers(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	users, err := app.users.GetAllSubscriptions(app.getUserId(r))
	if err != nil {
		log.Fatal(err)
	}
	nonUsers, err := app.users.GetAllNonSubscriptions(app.getUserId(r))
	if err != nil {
		log.Fatal(err)
	}
	data.Users = users
	data.NonUsers = nonUsers
	app.render(w, http.StatusOK, "recommend.html", data)
}

func (app *application) subscribe(w http.ResponseWriter, r *http.Request) {
	var form subForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	err = app.users.Subscribe(app.getUserId(r), form.Id)
	if err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/recommendations", http.StatusSeeOther)
}

func (app *application) unsubscribe(w http.ResponseWriter, r *http.Request) {
	var form subForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.users.Unsubscribe(app.getUserId(r), form.Id)
	if err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	user, err := app.users.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	posts, err := app.posts.GetAllMyPosts(id)
	if err != nil {
		log.Fatal(err)
	}

	users, err := app.users.GetAllSubscriptions(id)
	if err != nil {
		log.Fatal(err)
	}

	subs, err := app.users.GetAllSubscribers(id)
	if err != nil {
		log.Fatal(err)
	}

	data := app.newTemplateData(r)
	data.User = user
	data.Posts = posts
	data.Users = users
	data.Subs = subs

	app.render(w, http.StatusOK, "profile.html", data)
}
