all: pomo-server pomo

pomo-server: cmd/pomo-server/*.go
	go build -o bin/pomo-server ./cmd/pomo-server

pomo: cmd/pomo/*.go
	go build -o bin/pomo ./cmd/pomo

clean:
	rm -r bin/