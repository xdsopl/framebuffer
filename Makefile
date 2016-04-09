
.PHONY: all
all: $(addsuffix /fbgrab, 386 amd64 arm)

%/fbgrab: fbgrab.go
	GOARCH=$(@D) go build -o $@ $<

