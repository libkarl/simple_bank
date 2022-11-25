#Multistage build image for small image size

#Build stage
FROM golang:1.19.3-alpine3.16 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go
# The Alpine image doesn't have curl pre installed 
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz
#Run stage

FROM alpine:3.16
WORKDIR /app

# This will copy binary from first stage with name builder to /app directory as binary with name main
# The dot at the and says the place where this changes should start
COPY --from=builder /app/main .
# This line will copy binary for go migrate tool from builder stage
COPY --from=builder /app/migrate ./migrate
COPY app.env .
# copy script to docker image
COPY start.sh .
# This script is used for controling the order in which the services in
# docker compose are starting
COPY wait-for.sh .
# This line will copy db migration inside docker image
COPY db/migration ./migration

EXPOSE 8080

# This w
CMD ["/app/main"]
# definice entry pointu přikazu na start.sh script
# místo spuštění binárky main v app directory 
# se spustí scritp start.sh, který po provedení nadefinované sekvence operací
# spustí binárku main
ENTRYPOINT ["/app/start.sh"]