package core

import (
	"fmt"

	"github.com/demula/mono/example/api"
)

func Say(it api.Hello) string {
	return fmt.Sprintf("Hello %s", it.Who)
}

var you = api.Hello{
	Who: "you",
}

func SayYou() string {
	return fmt.Sprintf("Hello %s", you.Who)
}