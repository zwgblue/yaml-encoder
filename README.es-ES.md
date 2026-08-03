# yaml-encoder

Este es un codificador YAML que implementa `yaml.Marshaler` para marshal con comentarios.

Inspirado en [el codificador del documento de talos](https://github.com/siderolabs/talos/blob/main/pkg/machinery/config/encoder/encoder.go)

## ejemplo

```go
package main

import (
  "fmt"
  encoder "github.com/zwgblue/yaml-encoder"
)

type DBConfig struct {
  Username string `comment:"esto es el nombre de usuario de la base de datos"`
  Password string `comment:"esto es la contraseña de la base de datos"`
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

salida:

```yaml
# esto es el nombre de usuario de la base de datos
username: root
# esto es la contraseña de la base de datos
password: xxxxxx
```

Si no te gustaría usar la etiqueta `comment`, podrías usar `WithCustomizedTag` para cambiar:

```go
encoder := encoder.NewEncoder(config, encoder.WithCustomizedTag("yourTagName"))
```
