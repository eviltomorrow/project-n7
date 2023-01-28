FROM golang As builder
WORKDIR /project-n7
COPY [".", "./"]
ARG APPNAME=unknown
ARG MAINVERSION=unknown
ARG GITSHA=unknown
ARG BUILDTIME=unknown
ENV MAINVERSION=${MAINVERSION} \
    GITSHA=${GITSHA} \
    BUILDTIME=${BUILDTIME}
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X ain.MainVersion=${MAINVERSION} -X main.GitSha=${GITSHA} -X main.BuildTime=${BUILDTIME} -s -w" -gcflags "all=-trimpath=$(go env GOPATH)" -o bin/${APPNAME} app/${APPNAME}/main.go

FROM alpine:latest as prod
WORKDIR /app
COPY --from builder ["/project-n7/bin/${APPNAME}","./bin/"]
COPY --from builder ["/project-n7/app/${APPNAME}/etc","./etc/"]

