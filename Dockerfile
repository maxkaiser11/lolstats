# ---- Stage 1: build -----
FROM golang:1.27-alpine AS build
WORKDIR /src

# Copying only the dependencies first
COPY go.mod go.sum ./
RUN go mod download

# copying the rest of the code and compiling it
COPY . . 
RUN CGO_ENABLED=0 go build -o /lolstats ./cmd/api


# ---- Stage 2: run -----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /lolstats /lolstats
EXPOSE 8080
ENTRYPOINT [ "/lolstats" ]

