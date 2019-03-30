FROM golang:1.11 as builder
WORKDIR /go/src/github.com/codeboten/redistributed/
COPY vendor ./vendor
COPY propagation ./propagation
WORKDIR /go/src/github.com/codeboten/redistributed/api
COPY ./api/api.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/codeboten/redistributed/api/app .
CMD ["./app"]
