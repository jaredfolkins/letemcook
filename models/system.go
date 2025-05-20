package models

import (
	"context"
	"gopkg.in/yaml.v3"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/jaredfolkins/letemcook/util"
)

// CollectImages gathers unique container images used in all cookbooks and apps.
func CollectImages() ([]string, error) {
	cbs, err := AllCookbooks()
	if err != nil {
		return nil, err
	}
	apps, err := AllApps()
	if err != nil {
		return nil, err
	}

	images := map[string]struct{}{}

	extract := func(yml string) {
		if yml == "" {
			return
		}
		var yd YamlDefault
		if err := yaml.Unmarshal([]byte(yml), &yd); err == nil {
			for _, p := range yd.Cookbook.Pages {
				for _, r := range p.Recipes {
					for _, s := range r.Steps {
						if s.Image != "" {
							images[s.Image] = struct{}{}
						}
					}
				}
			}
		}
	}

	for _, cb := range cbs {
		extract(cb.YamlShared)
		extract(cb.YamlIndividual)
	}
	for _, ap := range apps {
		extract(ap.YAMLShared)
		extract(ap.YAMLIndividual)
	}

	list := make([]string, 0, len(images))
	for img := range images {
		list = append(list, img)
	}
	sort.Strings(list)
	return list, nil
}

// CollectImageInfos gathers metadata about each unique image used by cookbooks and apps.
func CollectImageInfos() ([]ImageInfo, error) {
	names, err := CollectImages()
	if err != nil {
		return nil, err
	}

	cli, err := client.NewClientWithOpts(
		client.WithHost(os.Getenv("LEMC_DOCKER_HOST")),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}

	infos := make([]ImageInfo, 0, len(names))

	for _, name := range names {
		info := ImageInfo{Name: name}
		normalized, _, _, err := util.NormalizeImageName(name)
		if err == nil {
			inspect, _, ierr := cli.ImageInspectWithRaw(context.Background(), normalized)
			if ierr == nil {
				info.Exists = true
				if !inspect.Metadata.LastTagTime.IsZero() {
					info.LastUpdated = inspect.Metadata.LastTagTime
				} else if t, perr := time.Parse(time.RFC3339Nano, inspect.Created); perr == nil {
					info.LastUpdated = t
				}
				localDigest := strings.TrimPrefix(inspect.ID, "sha256:")
				if remoteDigest, derr := getRemoteDigest(cli, normalized); derr == nil && remoteDigest != "" {
					if !strings.HasPrefix(localDigest, remoteDigest) {
						info.NewerAvailable = true
					}
				}
			}
		}
		infos = append(infos, info)
	}

	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })
	return infos, nil
}

func getRemoteDigest(cli *client.Client, imageRef string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	dist, err := cli.DistributionInspect(ctx, imageRef, "")
	if err != nil {
		if errdefs.IsNotFound(err) || errdefs.IsUnauthorized(err) {
			return "", nil
		}
		return "", err
	}
	if dist.Descriptor.Digest == "" {
		return "", nil
	}
	return string(dist.Descriptor.Digest), nil
}
