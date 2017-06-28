package workflows

import (
	"archive/zip"
	"fmt"
	"github.com/stelligent/mu/common"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// NewServicePusher create a new workflow for pushing a service to a repo
func NewServicePusher(ctx *common.Context, tag string, provider string, dockerWriter io.Writer) Executor {

	workflow := new(serviceWorkflow)

	return newPipelineExecutor(
		workflow.serviceLoader(ctx, tag, provider),
		newConditionalExecutor(workflow.isEcrProvider(),
			newPipelineExecutor(
				workflow.serviceRepoUpserter(&ctx.Config.Service, ctx.StackManager, ctx.StackManager),
				workflow.serviceImageBuilder(ctx.DockerManager, &ctx.Config, dockerWriter),
				workflow.serviceRegistryAuthenticator(ctx.ClusterManager),
				workflow.serviceImagePusher(ctx.DockerManager, dockerWriter),
			),
			newPipelineExecutor(
				workflow.serviceBucketUpserter(&ctx.Config.Service, ctx.StackManager, ctx.StackManager),
				workflow.serviceArchiveUploader(ctx.Config.Basedir, ctx.ArtifactManager),
			)))

}

func (workflow *serviceWorkflow) serviceImageBuilder(imageBuilder common.DockerImageBuilder, config *common.Config, dockerWriter io.Writer) Executor {
	return func() error {
		log.Noticef("Building service:'%s' as image:%s'", workflow.serviceName, workflow.serviceImage)
		return imageBuilder.ImageBuild(config.Basedir, config.Service.Dockerfile, []string{workflow.serviceImage}, dockerWriter)
	}
}

func (workflow *serviceWorkflow) serviceImagePusher(imagePusher common.DockerImagePusher, dockerWriter io.Writer) Executor {
	return func() error {
		log.Noticef("Pushing service '%s' to '%s'", workflow.serviceName, workflow.serviceImage)
		return imagePusher.ImagePush(workflow.serviceImage, workflow.registryAuth, dockerWriter)
	}
}

func (workflow *serviceWorkflow) serviceArchiveUploader(basedir string, artifactCreator common.ArtifactCreator) Executor {
	return func() error {
		destURL := fmt.Sprintf("s3://%s/%s/%s.zip", workflow.serviceBucket, workflow.serviceName, workflow.serviceTag)
		log.Noticef("Pushing archive '%s' to '%s'", basedir, destURL)

		zipfile, err := zipDir(basedir)
		if err != nil {
			return err
		}
		defer os.Remove(zipfile.Name()) // clean up

		err = artifactCreator.CreateArtifact(zipfile, destURL)
		if err != nil {
			return err
		}

		return nil
	}
}

func zipDir(basedir string) (*os.File, error) {
	zipfile, err := ioutil.TempFile("", "artifact")
	if err != nil {
		return nil, err
	}

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	log.Debugf("Creating zipfile '%s' from basedir '%s'", zipfile, basedir)

	filepath.Walk(basedir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		log.Debugf(" ..Adding file '%s'", path)

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return zipfile, err
}
