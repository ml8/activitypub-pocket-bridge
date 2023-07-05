GOOS=linux
GOARCH=amd64
APPNAME=ap-bot
BINNAME=$(APPNAME)-$(GOOS)-$(GOARCH)
VM=ap-dev
DEPS=activitypub/*.go pocket/*.go main.go util/*.go

.PHONY: pocket-env-valid docker-env-valid conainer-build deploy upload-remote run-local run-remote clean

include .env


$(APPNAME): $(DEPS)
	go build -o $(APPNAME) main.go

$(BINNAME): $(DEPS)
	env GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINNAME) main.go

upload-remote: $(BINNAME)
	gcloud compute scp $(BINNAME) $(VM):./

run-local: $(APPNAME)
	./$(APPNAME) --pocketAppKey=$(AP_POCKET_KEY) --logtostderr 

run-remote: upload
	gcloud compute ssh $(VM) -- sudo ./$(BINNAME) --prod --pocketAppKey=$(AP_POCKET_KEY) --logtostderr 

clean:
	rm -f $(APPNAME) $(BINNAME)
