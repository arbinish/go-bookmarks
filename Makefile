.PHONY: all clean

all: go-bookmarks bookmark


go-bookmarks:
	go build -o go-bookmarks

bookmark:
	go build -o bookmark cli/main.go

clean:
	rm -f go-bookmarks bookmark
