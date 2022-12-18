FROM golang:1.19-alpine as scratch

WORKDIR /src/app

COPY ./tester_service/tester_service.go .

EXPOSE 9010

CMD ["go", "run", "tester_service.go", "-service-api", "service:9090"]
