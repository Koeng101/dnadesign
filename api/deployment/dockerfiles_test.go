package dockerfiles_test

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

func TestApiDockerfile(t *testing.T) {
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()

	cli := provider.Client()
	ctx := context.Background()
	tag, err := provider.BuildImage(ctx, &testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../../.",
			Dockerfile: "api/Dockerfile",
			Repo:       "dnadesignapi",
			Tag:        "test",
		},
	})
	if err != nil {
		t.Errorf("BuildImage should be nil. Got err: %s", err)
	}
	if tag != "dnadesignapi:test" {
		t.Errorf("Improper tag set. \nGot: %s\nExpect:%s", tag, "dnadesignapi:test")
	}

	_, _, err = cli.ImageInspectWithRaw(ctx, tag)
	if err != nil {
		t.Errorf("ImageInspect should be nil. Got err: %s", err)
	}

	t.Cleanup(func() {
		_, err := cli.ImageRemove(ctx, tag, types.ImageRemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}
