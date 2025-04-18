# Build
FROM golang:alpine AS build
WORKDIR /build/
COPY go.mod .
RUN go mod tidy
COPY . .
RUN go build

# Production
FROM alpine AS production
WORKDIR /chrdsdatamanager/
COPY --from=build /build/chrdsdatamanager ./
COPY ./docs ./docs
COPY ./*.crt ./
COPY ./*.key ./
COPY ./*.pem ./
COPY ./chrdsdatamanager.json ./

EXPOSE 6006 10051

CMD ["./chrdsdatamanager", "-log=stdout"]