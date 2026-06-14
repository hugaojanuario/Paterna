package commands

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

var coisoCmd = &cobra.Command{
	Use:   "init",
	Short: "Configuração inicial",
	Run:   upDocker,
}

func init() {
	rootCmd.AddCommand(coisoCmd)
}

func upDocker(cmd *cobra.Command, args []string) {

	if !validateToken() {
		fmt.Println("use 'paterna auth --login' ou 'paterna auth --register'")
		return
	}

	fmt.Println("Subindo o docker...")
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		panic(err)
	}

	reader, err := cli.ImagePull(context.Background(), "https://hub.docker.com/r/devhugojanuario/paterna", image.PullOptions{})
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image: "Paterna:latest",
	}, nil, nil, nil, "Paterna")

	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println("Docker subiu com sucesso!")
}
