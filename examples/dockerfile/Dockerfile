FROM golang:1.13 as builder

RUN mkdir /build
COPY main.go /build/main.go

WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp .

FROM alpine:latest

ARG HOST
ENV MYAPP_HOST=${HOST}

COPY --from=builder /build/myapp /app/
RUN chmod +x /app/myapp
EXPOSE 8080

CMD /app/myapp