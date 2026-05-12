## ---- Frontend Build ----
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install --registry https://registry.npmmirror.com
COPY frontend/ ./
RUN npm run build

## ---- Backend Build ----
FROM golang:alpine AS backend-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=1 go build -o /app/aigate .

## ---- Runtime ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=backend-builder /app/aigate .
COPY --from=backend-builder /app/frontend/dist ./frontend/dist

EXPOSE 8081
ENTRYPOINT ["./aigate"]
