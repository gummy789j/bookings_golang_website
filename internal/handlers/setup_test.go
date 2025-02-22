package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gummy789j/bookings/internal/config"
	"github.com/gummy789j/bookings/internal/models"
	"github.com/gummy789j/bookings/internal/render"
	"github.com/justinas/nosurf"
)

var app config.AppConfig

var session *scs.SessionManager

var pathToTemplates = "../../templates"

var functions = template.FuncMap{

	"humanDate":  render.HumanDate,
	"formatDate": render.FormatDate,
	"iterate":    render.Iterate,
	"add":        render.Add,
}

func TestMain(m *testing.M) {

	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Reservation{})
	gob.Register(models.Restriction{})
	gob.Register(models.Room{})
	gob.Register(models.RoomRestriction{})
	gob.Register(map[string]int{})

	app.InProduction = false
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	app.MailChan = make(chan models.MailData)
	defer close(app.MailChan)

	ListenForMail()

	tc, err := CreateTestTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
	}

	app.TemplateCache = tc
	app.UseCache = true

	repo := NewTestRepo(&app)
	NewHandlers(repo)

	render.NewRenderer(&app)

	os.Exit(m.Run())
}

func ListenForMail() {
	go func() {
		for {
			_ = <-app.MailChan
		}
	}()
}

func getRoutes() http.Handler {

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)

	//mux.Use(WriteToConsole)
	//mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", Repo.Home)
	mux.Get("/about", Repo.About)
	mux.Get("/search-availability", Repo.Availability)
	mux.Post("/search-availability", Repo.PostAvailability)
	mux.Get("/make-reservation", Repo.Reservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/generals-quarters", Repo.Generals)
	mux.Get("/majors-suite", Repo.Majors)
	mux.Get("/contact", Repo.Contact)
	mux.Post("/search-availability-json", Repo.JsonAvailability)
	mux.Get("/reservation-summary", Repo.ReservationSummary)
	mux.Get("/user/logout", Repo.Logout)
	mux.Get("/user/login", Repo.ShowLogin)
	mux.Post("/user/login", Repo.PostShowLogin)
	mux.Get("/admin/dashboard", Repo.AdminDashBoard)
	mux.Get("/admin/reservations-new", Repo.AdminNewReservations)
	mux.Get("/admin/reservations-all", Repo.AdminAllReservations)
	mux.Get("/admin/reservations-calendar", Repo.AdminReservationsCalendar)
	mux.Post("/admin/reservations-calendar", Repo.AdminPostReservationsCalendar)
	mux.Get("/admin/reservations/{src}/{id}/show", Repo.AdminShowReservation)
	mux.Post("/admin/reservations/{src}/{id}", Repo.AdminPostShowReservation)
	mux.Get("/admin/process-reservations/{src}/{id}/do", Repo.AdminProcessReservation)
	mux.Get("/admin/delete-reservations/{src}/{id}/do", Repo.AdminDeleteReservation)

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}

// NoSurf : adds CSRF protection to all POST requests
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

// SessionLoad : loads and saves the session on every request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// *template.Template是一個解析過後的html(...等)的file，也就是一些儲存text的fragment的在的記憶體位置
func CreateTestTemplateCache() (map[string]*template.Template, error) {

	myCache := make(map[string]*template.Template)

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	for _, page := range pages {

		name := filepath.Base(page)

		// 一個Tempalte重要的包含物 Name & content
		// Template是一個定義好的struct 裡面包含 Tree struct 跟 nameSpace
		// Tree 的 型別是 *parse.Tree 是定義在parse package中，
		// 我們New一個Template，傳進去的name就是存在Tree.Name中，作為這個parse file的名字
		// 而 nameSpace對應到的就是associate file 也就是我們最尾端呼叫的 ParseFile method所傳入的我們"原本的"html file
		// 要做的就是重新建立一個Template然後幫這個template加入一個function map(讓以後擴展性更高)，再把原內容加入進去
		// 這麼做的目的有3個
		// 1.為了讓他快速讀取修改後的html file(不然都要重新run程式開socket)
		// 2.也包含將layout的定義和內容加入page中
		// 3.可以自訂義新的tempalte的function(增加靈活性)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}
		//fmt.Println(ts)

		// ts, err := template.ParseFiles(page)
		// if err != nil {
		// 	return myCache, err
		// }

		// matches, err := filepath.Glob("./templates/*.layout.tmpl")
		// if err != nil {
		// 	return myCache, err
		// }

		// if len(matches) > 0 {
		// 	ts, err = ts.ParseGlob("./templates/*.layout.tmpl")
		// 	if err != nil {
		// 		return myCache, err
		// 	}
		// }

		ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		myCache[name] = ts

	}

	return myCache, nil

}
