FROM golang:latest
WORKDIR go/src/github.com/wwgberlin/go-event-sourcing-exercise
COPY . .
RUN go get github.com/notnil/chess && \
    go build -o app .
CMD ./app
EXPOSE 8080
