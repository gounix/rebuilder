ARG GOVERSION
ARG REBUILDER_VERSION

FROM docker.io/library/golang:${GOVERSION}-alpine AS build
ARG GOVERSION
ARG REBUILDER_VERSION
WORKDIR /
COPY ./ /

RUN go mod edit -C main -go $GOVERSION
RUN go mod tidy -C main
RUN sed -i s/Development-version/${REBUILDER_VERSION}/ /main/main.go
RUN sed -i s/Golang-version/${GOVERSION}/ /main/main.go

RUN go build -C main -o rebuilder main.go timer.go

FROM docker.io/library/golang:${GOVERSION}-alpine
ARG GOVERSION

COPY --from=build ./main/rebuilder /rebuilder

WORKDIR /
CMD [ "/rebuilder" ]

