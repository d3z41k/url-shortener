package main

import (
	"fmt"
	h "github.com/d3z41k/url-shortener/api"
	mor "github.com/d3z41k/url-shortener/repository/mongo"
	myr "github.com/d3z41k/url-shortener/repository/mysql"
	rr "github.com/d3z41k/url-shortener/repository/redis"
	slr "github.com/d3z41k/url-shortener/repository/sqlite"
	"github.com/d3z41k/url-shortener/shortener"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

// https://www.google.com -> 98sj1-293
// http://localhost:8000/98sj1-293 -> https://www.google.com

// repo <- service -> serializer  -> http

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	repo := chooseRepo()
	service := shortener.NewRedirectService(repo)
	handler := h.NewHandler(service)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/{code}", handler.Get)
	r.Post("/", handler.Post)

	errs := make(chan error, 2)
	go func() {
		fmt.Println("Listening on port :8000")
		errs <- http.ListenAndServe(httpPort(), r)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	fmt.Printf("Terminated %s", <-errs)
}

func httpPort() string {
	port := "8000"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	return fmt.Sprintf(":%s", port)
}

func chooseRepo() shortener.RedirectRepository {
	switch os.Getenv("URL_DB") {
	case "redis":
		redisURL := os.Getenv("REDIS_URL")
		repo, err := rr.NewRedisRepository(redisURL)
		if err != nil {
			log.Fatal(err)
		}
		return repo
	case "mongo":
		mongoURL := os.Getenv("MONGO_URL")
		mongodb := os.Getenv("MONGO_DB")
		mongoTimeout, _ := strconv.Atoi(os.Getenv("MONGO_TIMEOUT"))
		repo, err := mor.NewMongoRepository(mongoURL, mongodb, mongoTimeout)
		if err != nil {
			log.Fatal(err)
		}
		return repo
	case "sqlite":
		repo, err := slr.NewSqliteRepository(os.Getenv("SQLITE_DB"))
		if err != nil {
			log.Fatal(err)
		}
		return repo
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?&charset=utf8&interpolateParams=true",
			os.Getenv("MYSQL_USER"),
			os.Getenv("MYSQL_PASSWORD"),
			os.Getenv("MYSQL_HOST"),
			os.Getenv("MYSQL_PORT"),
			os.Getenv("MYSQL_NAME"))

		repo, err := myr.NewMysqlRepository(dsn)
		if err != nil {
			log.Fatal(err)
		}
		return repo
	}
	return nil
}
