package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gummy789j/bookings/internal/config"
	"github.com/gummy789j/bookings/internal/handlers"
)

func routes(app *config.AppConfig) http.Handler {

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)

	//mux.Use(WriteToConsole)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)
	//mux.Use(Auth)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/search-availability", handlers.Repo.Availability)
	mux.Post("/search-availability", handlers.Repo.PostAvailability)
	mux.Get("/make-reservation", handlers.Repo.Reservation)
	mux.Post("/make-reservation", handlers.Repo.PostReservation)
	mux.Get("/generals-quarters", handlers.Repo.Generals)
	mux.Get("/majors-suite", handlers.Repo.Majors)
	mux.Get("/contact", handlers.Repo.Contact)
	mux.Post("/search-availability-json", handlers.Repo.JsonAvailability)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)
	mux.Get("/choose-room/{id}", handlers.Repo.ChooseRoom)
	mux.Get("/book-room", handlers.Repo.BookRoom)
	mux.Get("/user/login", handlers.Repo.ShowLogin)
	mux.Get("/user/logout", handlers.Repo.Logout)
	mux.Post("/user/login", handlers.Repo.PostShowLogin)
	mux.Get("/login", handlers.Repo.ShowLogin)
	mux.Get("/logout", handlers.Repo.Logout)
	mux.Post("/login", handlers.Repo.PostShowLogin)

	// 只要有Open method的都屬於FileSystem interface
	// type Dir string 就有Open method
	// Dir() 更像是一種強制轉型
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	mux.Route("/admin", func(mux chi.Router) {
		//mux.Use(Auth)
		mux.Get("/dashboard", handlers.Repo.AdminDashBoard)
		mux.Get("/reservations-new", handlers.Repo.AdminNewReservations)
		mux.Get("/reservations-all", handlers.Repo.AdminAllReservations)
		mux.Get("/reservations-calendar", handlers.Repo.AdminReservationsCalendar)
		mux.Post("/reservations-calendar", handlers.Repo.AdminPostReservationsCalendar)
		mux.Get("/reservations/{src}/{id}/show", handlers.Repo.AdminShowReservation)
		mux.Post("/reservations/{src}/{id}", handlers.Repo.AdminPostShowReservation)
		mux.Get("/process-reservations/{src}/{id}/do", handlers.Repo.AdminProcessReservation)
		mux.Get("/delete-reservations/{src}/{id}/do", handlers.Repo.AdminDeleteReservation)

	})

	return mux
}
