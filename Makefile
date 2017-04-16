
.PHONY: all
all: $(addsuffix /fbshow, 386 amd64 arm) $(addsuffix /fbgrab, 386 amd64 arm) $(addsuffix /fbtouch, 386 amd64 arm)

%/fbgrab: fbgrab.go
	GOARCH=$(@D) GOPATH=$(CURDIR) go build -ldflags -s -o $@ $<

%/fbshow: fbshow.go
	GOARCH=$(@D) GOPATH=$(CURDIR) go build -ldflags -s -o $@ $<

%/fbtouch: fbtouch.go
	GOARCH=$(@D) GOPATH=$(CURDIR) go build -ldflags -s -o $@ $<

