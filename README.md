# go-web-test
A wiki web app start with https://golang.org/doc/articles/wiki/

1. write a simple middleware framework https://github.com/withwind8/middleware
2. using jteeuwen/go-bindata to embedding tmpl
3. using gorilla/mux as router

## Usage
```bash
go get -d github.com/withwind8/go-web-test  #or git clone
go get github.com/jteeuwen/go-bindata/...

# go to repo dir
go generate
go build
./go-web-test
```

Now, visit http://127.0.0.1:8080

Options of go-web-test:
```
  -addr string
    	Server listen addr (default "127.0.0.1")
  -data string
    	Where wiki data store (default "./data")
  -port int
    	Server listen port (default 8080)
```
