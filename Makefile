build_path = bin
entrypoint = cmd/web/main.go

build:
	go build -ldflags="-s -w" -o $(build_path)/URLRotator.exe $(entrypoint)

run:
	go run $(entrypoint)