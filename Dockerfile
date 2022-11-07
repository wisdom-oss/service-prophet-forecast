FROM golang:buster AS build-service
COPY src /tmp/src
WORKDIR /tmp/src
RUN mkdir -p /tmp/build
RUN go mod download
RUN go build -o /tmp/build/app

FROM rocker/r-ver:latest
COPY --from=build-service /tmp/build/app /service
COPY res /res
RUN Rscript /res/packages.r
WORKDIR /
ENTRYPOINT ["/service"]
HEALTHCHECK --interval=10s CMD /service -healthcheck
