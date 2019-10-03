package.zip: forvo.exe
	rm -fr forvo
	mkdir forvo
	cp forvo.* forvo/
	zip forvo.zip -r forvo

forvo.exe: main.go
	GOOS=windows GOARCH=amd64 go build

.PHONY: clean

clean:
	rm -f forvo.exe
	rm -f forvo.zip
	rm -fr forvo
