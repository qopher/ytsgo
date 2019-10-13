# ytsgo
Go library to access YTS.LT API

API documentation: https://yts.lt/api

## Example
```go
  c, err := ytsgo.New()
  if err != nil {
    log.Fatalf("Failed to create ytsgo client: %v", err)
  }
  m, err := c.Movie(10)
  if err != nil {
    log.Fatalf("Failed to fetch movie id:%v :%v", id, err)
  }
  fmt.Println(m.Title)
```

[ytsgo.go](https://github.com/qopher/ytsgo/blob/master/cmd/ytsgo.go) is a simple CLI tool to fetch and search movies from YTS.LG and show manget links to all torrents.
