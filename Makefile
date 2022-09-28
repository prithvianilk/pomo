all: server pomo

server: cmd/server/*.go
	go build -o bin/server ./cmd/server

pomo: cmd/pomo/*.go
	go build -o bin/pomo ./cmd/pomo

clean:
	rm -r bin/