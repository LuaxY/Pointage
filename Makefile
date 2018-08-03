all: prepare build

get:
	set GOOS=windows && set GOARCH=386 && go get

prepare:
	windres -o pointage.syso pointage.rc

build:
	set GOOS=windows
	set GOARCH=386
	go build --ldflags="-s -w -H windowsgui" -o Pointage.exe