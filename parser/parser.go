package parser

import (
	"context"
	"fmt"
	"io"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

type CronJobParser struct {
	logger *zap.Logger
}

func ProvideCronJobParser(
	ctx context.Context,
	logger *zap.Logger,
) (*CronJobParser, error) {
	return &CronJobParser{
		logger: logger,
	}, nil
}

func (p *CronJobParser) ParseCronJobConfigs(
	r io.ReadCloser,
) []batchv1.CronJob {
	cronJobs := []batchv1.CronJob{}
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
			p.logger.Sugar().Error(
				"failed to parse yaml config: ",
				err,
			)
		}

		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}
		if unstructuredObj.GetKind() == "CronJob" {
			var cronJob batchv1.CronJob
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(
				unstructuredMap,
				&cronJob,
			)
			if err != nil {
				p.logger.Sugar().Error(
					"failed to parse cronjob: ",
					err,
				)
				continue
			}
			fmt.Println(cronJob.Name)
			cronJobs = append(cronJobs, cronJob)
		}
	}
	if err != io.EOF {
		p.logger.Sugar().Error(
			"expected EOF got: ",
			err,
		)
	}

	return cronJobs
}
