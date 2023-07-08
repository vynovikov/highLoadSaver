FROM golang:1.21-rc-bullseye as build

WORKDIR /highLoadSaver

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o highLoadSaver ./cmd/highLoadSaver

CMD ./highLoadSaver

FROM alpine:latest as release

RUN apk --no-cache add ca-certificates && \
	mkdir /tls


COPY --from=build /highLoadSaver ./ 

RUN chmod +x ./highLoadSaver

ENTRYPOINT [ "./highLoadSaver" ]