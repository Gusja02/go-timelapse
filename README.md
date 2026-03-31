
docker login

docker build -t gusja/go-timelapse:v1.0 .

docker compose -f local.docker-compose.yml build
docker tag go-timelapse:latest gusja/go-timelapse:v1.0

docker buildx build --platform linux/amd64,linux/arm64 -t gusja/go-timelapse:v1.0 --push .
