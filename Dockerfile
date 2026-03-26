# Dockerfile
FROM node:20-alpine AS frontend-build
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npx nuxt generate

FROM golang:1.22-alpine AS backend-build
RUN apk add --no-cache gcc musl-dev
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend-build /app/frontend/.output/public ./static/
RUN CGO_ENABLED=1 go build -o cronjob-panel .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend-build /app/backend/cronjob-panel .
RUN mkdir -p data
EXPOSE 8080
CMD ["./cronjob-panel"]
