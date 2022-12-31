FROM golang:1.19-alpine AS build

WORKDIR /app

COPY src/go.mod ./
COPY src/go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-traitor


##
## Deploy
##

FROM scratch

WORKDIR /

COPY --from=build /docker-traitor /docker-traitor

ENTRYPOINT ["/docker-traitor"]