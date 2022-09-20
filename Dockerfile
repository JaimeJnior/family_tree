# syntax=docker/dockerfile:1

FROM golang:1.19.1-alpine

WORKDIR /family_tree_app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build -o ./out/family-tree-app ./cmd

EXPOSE 8080

CMD [ "./out/family-tree-app" ]