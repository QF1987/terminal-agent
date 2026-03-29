.PHONY: build run clean test

build:
	go build -o bin/device-ctl ./cmd/device-ctl

run: build
	./bin/device-ctl $(ARGS)

clean:
	rm -rf bin/

test:
	go test ./...

# 快捷命令
list: build
	./bin/device-ctl list

list-offline: build
	./bin/device-ctl list --status offline

list-region: build
	./bin/device-ctl list --region $(REGION)

info: build
	./bin/device-ctl info $(ID)

stats: build
	./bin/device-ctl stats $(ID)

reboot: build
	./bin/device-ctl reboot $(ID)

logs: build
	./bin/device-ctl logs

alerts: build
	./bin/device-ctl monitor alerts

status: build
	./bin/device-ctl monitor status
