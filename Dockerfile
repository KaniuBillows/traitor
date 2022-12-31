FROM golang:1.19-alpine AS build

WORKDIR /app

COPY src ./



RUN go env -w  GOPROXY=https://goproxy.io,direct

RUN go mod download


RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-traitor
COPY src/ui /ui

##
## Deploy
##

FROM scratch

WORKDIR /

# copy exe
COPY --from=build /docker-traitor /docker-traitor
COPY --from=build /ui /ui

ENTRYPOINT ["/docker-traitor"]