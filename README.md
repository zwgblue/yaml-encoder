# yaml-encoder

This is a yaml encoder which implements `yaml.Marshaler` for marshal with comments.

Inspired by [The doc's encoder of talos](https://github.com/siderolabs/talos/blob/main/pkg/machinery/config/encoder/encoder.go)

## example

```go
package main

import (
  "fmt"
  encoder "github.com/zwgblue/yaml-encoder"
)

type DBConfig struct {
  Username string `comment:"this is the username of database"`
  Password string `comment:"this is the password of database"`
}

func main() {
    config := DBConfig{
      Username: "root",
      Password: "xxxxxx",
    }

    encoder := encoder.NewEncoder(config, encoder.WithComments(encoder.CommentsOnHead))
    content, _ := encoder.Encode()
    fmt.Printf("%s", content)
}
```

output:

```yaml
# this is the username of database
username: root
# this is the password of database
password: xxxxxx
```

If you don't like to use the `comment` tag, you could use the `WithCustomizedTag` to change:

```go
encoder := encoder.NewEncoder(config, encoder.WithCustomizedTag("yourTagName"))
```
