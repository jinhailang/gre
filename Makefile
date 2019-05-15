NAME=gre
INSTALL_PATH=/opt/bin/

build:
	go build -race -ldflags '-w -s' -o $(NAME) *.go

test:
	go test -v -cover

install:
	make build && mv ./$(NAME) $(INSTALL_PATH)

