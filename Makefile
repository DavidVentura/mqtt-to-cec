.PHONY: all

all: mqtt_arm64 mqtt_amd64

mqtt_arm64:
	GOARCH=arm64 go build -o mqtt_arm64 -ldflags="-s -w"  mqtt.go
mqtt_amd64:
	GOARCH=amd64 go build -o mqtt_amd64 -ldflags="-s -w"  mqtt.go
