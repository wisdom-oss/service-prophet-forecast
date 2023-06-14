module microservice

go 1.19

require github.com/sirupsen/logrus v1.9.0

require (
	github.com/go-chi/chi/v5 v5.0.8
	github.com/go-chi/httplog v0.2.5
	github.com/gosimple/slug v1.13.1
	github.com/lib/pq v1.10.6
	github.com/qustavo/dotsql v1.1.0
	github.com/rs/zerolog v1.27.0
	github.com/wisdom-oss/commonTypes v1.0.0
	github.com/wisdom-oss/golang-kong-access v0.2.2
	github.com/wisdom-oss/microservice-middlewares/v2 v2.0.0-20230607132432-2c53e4e04ce8
)

require (
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/wisdom-oss/microservice-utils v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20221006211917-84dc82d7e875 // indirect
)
