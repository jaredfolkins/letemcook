package models

import (
	"gopkg.in/yaml.v3"
	"sort"
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
