FROM golang:alpine as builder
LABEL maintainer="Chris Torstenson <chris.torstenson@gmail.com>"
ENV PORT=8083
RUN apk update && apk add --no-cache git
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -mod=readonly -v -o main .
RUN apk --no-cache add ca-certificates gcc g++ make
EXPOSE 8083
CMD ["./main"]
