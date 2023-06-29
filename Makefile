GOOS=linux
GOARCH=amd64
APPNAME=ap-bot
BINNAME=$(APPNAME)-$(GOOS)-$(GOARCH)
VM=ap-dev

crossbuild:
	env GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BINNAME) main.go

build:
	go build -o ap-bot main.go

upload: crossbuild
	gcloud compute scp $(BINNAME) $(VM):./

local:
	./$(APPNAME) --pocketAppKey=$(POCKET_KEY) --logtostderr --initUser=$(POCKET_USER) --initTok=$(POCKET_USER_TOKEN)

run: upload
	gcloud compute ssh $(VM) -- sudo ./$(BINNAME) --prod --pocketAppKey=$(POCKET_KEY) --logtostderr --initUser=$(POCKET_USER) --initTok=$(POCKET_USER_TOKEN)

clean:
	rm -f $(APPNAME) $(BINNAME)
