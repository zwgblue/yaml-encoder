package yaml_encoder_test

import (
	"fmt"

	encoder "github.com/zwgblue/yaml-encoder"
)

func ExampleEncoder() {
	type DBConfig struct {
		Username string `comment:"this is the username of database"`
		Password string `comment:"this is the password of database"`
	}

	config := DBConfig{
		Username: "root",
		Password: "xxxxxx",
	}

	encoder := encoder.NewEncoder(config, encoder.WithComments(encoder.CommentsOnHead))
	content, _ := encoder.Encode()
	fmt.Printf("%s", content)
	// Output:
	// # this is the username of database
	// username: root
	// # this is the password of database
	// password: xxxxxx
}
