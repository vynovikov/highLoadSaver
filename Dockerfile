FROM golang:1.20-buster as build

WORKDIR /postSaver

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o postSaver ./cmd/postSaver

#CMD ./postSaver

FROM alpine:latest as release

RUN apk --no-cache add ca-certificates

COPY --from=build /postSaver ./ 

RUN chmod +x ./postSaver

ENTRYPOINT [ "./postSaver" ]

EXPOSE 3100