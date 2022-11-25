#Multistage build image for small image size

#Build stage
FROM golang:1.19.3-alpine3.16 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

#Run stage

FROM alpine:3.16
WORKDIR /app

# This will copy binary from first stage with name builder to /app directory as binary with name main
# The dot at the and says the place where this changes should start
COPY --from=builder /app/main .
COPY app.env .

EXPOSE 8080
CMD ["/app/main"]