package docker

import (
	"sync"

	"github.com/docker/docker/client"
)

var (
	instance *client.Client
	initErr  error
	once     sync.Once
)

// GetClient devolve um cliente Docker compartilhado, criado uma única vez
// a partir do ambiente (DOCKER_HOST etc) com negociação de versão da API.
func GetClient() (*client.Client, error) {
	once.Do(func() {
		instance, initErr = client.NewClientWithOpts(
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		)
	})
	return instance, initErr
}
