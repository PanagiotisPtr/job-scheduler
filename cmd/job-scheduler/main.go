package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/google/go-github/v48/github"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

func IsYaml(path string) bool {
	return strings.Contains(path, ".yml") ||
		strings.Contains(path, ".yaml")
}

type RepositoryArgs struct {
	Owner string
	Name  string
	Path  string
}

func getYmlFilePaths(
	ctx context.Context,
	client *github.Client,
	args RepositoryArgs,
) ([]RepositoryArgs, error) {
	rv := []RepositoryArgs{}
	paths := []string{args.Path}

	for len(paths) > 0 {
		p := paths[len(paths)-1]
		paths = paths[:len(paths)-1]

		_, content, _, err := client.Repositories.GetContents(
			ctx,
			args.Owner,
			args.Name,
			p,
			nil,
		)
		if err != nil {
			// will log this later
			continue
		}

		for _, c := range content {
			switch c.GetType() {
			case "dir":
				if c.Path != nil {
					paths = append(paths, c.GetPath())
				}
			case "file":
				fmt.Println(c.GetContent())
				if IsYaml(c.GetPath()) {
					rv = append(rv, RepositoryArgs{
						Owner: args.Owner,
						Name:  args.Name,
						Path:  c.GetPath(),
					})
				}
			}
		}
	}

	return rv, nil
}

func GetFile(
	ctx context.Context,
	client *github.Client,
	args RepositoryArgs,
) error {
	r, _, err := client.Repositories.DownloadContents(
		ctx,
		args.Owner,
		args.Name,
		args.Path,
		nil,
	)
	if err != nil {
		return err
	}
	defer r.Close()

	ParseCronJobConfigs(r)

	return nil
}

func ParseCronJobConfigs(
	r io.ReadCloser,
) []*batchv1.CronJob {
	cronJobs := []*batchv1.CronJob{}
	var err error
	decoder := yamlutil.NewYAMLOrJSONDecoder(r, 100)
	for {
		var rawObj runtime.RawExtension
		if err = decoder.Decode(&rawObj); err != nil {
			break
		}

		obj, _, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			log.Fatal(err)
		}

		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

		if unstructuredObj.GetKind() == "CronJob" {
			var cronJob batchv1.CronJob
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(
				unstructuredMap,
				&cronJob,
			)
			if err != nil {
				continue
			}
			fmt.Println(cronJob.Name)
			cronJobs = append(cronJobs, &cronJob)
		}
	}
	if err != io.EOF {
		log.Fatal("eof ", err)
	}

	return cronJobs
}

func main() {
	ctx := context.Background()
	client := github.NewClient(nil)

	paths, err := getYmlFilePaths(
		ctx,
		client,
		RepositoryArgs{
			Owner: "panagiotisptr",
			Name:  "hermes-messenger",
			Path:  "/deployment",
		},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(paths)

	for _, f := range paths {
		err = GetFile(ctx, client, f)
	}
}
