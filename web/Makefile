.PHONY: all build clean dev

# Build the complete server binary with embedded frontend
all: build

build: frontend server

frontend:
	bun run build
	rm -rf server/dist
	cp -r dist server/dist

server:
	cd server && go build -o ../facturx-server .

# Development mode (frontend + backend separately)
dev:
	@echo "Start backend: cd server && go run ."
	@echo "Start frontend: npm run dev"

clean:
	rm -rf dist server/dist facturx-server

# Build for Linux (for deployment)
build-linux:
	npm run build
	rm -rf server/dist
	cp -r dist server/dist
	cd server && GOOS=linux GOARCH=amd64 go build -o ../facturx-server .
