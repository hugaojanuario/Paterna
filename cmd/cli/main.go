package main

import (
	"github.com/hugaojanuario/Paterna/internal/commands"
	"github.com/hugaojanuario/Paterna/internal/repository"
	"github.com/hugaojanuario/Paterna/pkg/dotenv"
)

// @ViitoJooj:
// fun fact, i love boostrap!!! S2

// @Bugo:
// rubber duck
// rubber duck
// rubber duck
// rubber duck
// rubber duck
// rubber duck
// rubber duck
// rubber duck
// rubber duck
// rubber duck
// rubber duck
func main() {
	dotenv.Catch()
	repository.Init()
	commands.Execute()
}
