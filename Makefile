default: gojson

gojson:
	go build -o bin

install: gojson
	chmod a+x bin/gojson
	install bin/gojson ~/opt/bin