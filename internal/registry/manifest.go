// Extracted from https://containrrr.dev/watchtower/

package registry

import (
	"fmt"
	"net/url"
	"strings"

	ref "github.com/distribution/reference"
)

func getHost(image string) (string, error) {
	normalizedName, err := ref.ParseNormalizedNamed(image)
	if err != nil {
		return "", err
	}

	host, err := NormalizeRegistry(normalizedName.String())
	if err != nil {
		return "", err
	}

	return host, nil
}

// BuildManifestURL from raw image data
func BuildManifestURL(image string) (string, error) {
	host, err := getHost(image)
	if err != nil {
		return "", err
	}

	img, tag := extractImageAndTag(strings.TrimPrefix(image, host+"/"))
	img = GetScopeFromImageName(img, host)
	if !strings.Contains(img, "/") {
		img = "library/" + img
	}

	url := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   fmt.Sprintf("/v2/%s/manifests/%s", img, tag),
	}

	return url.String(), nil
}

// extractImageAndTag from a concatenated string
func extractImageAndTag(imageName string) (string, string) {
	if imageName == "" {
		return "", "latest"
	}

	if strings.Contains(imageName, ":") {
		parts := strings.Split(imageName, ":")
		if len(parts) == 0 {
			return "", "latest"
		}
		if len(parts) == 1 {
			return parts[0], "latest"
		}
		if len(parts) == 2 {
			return parts[0], parts[1]
		}

		return parts[0], strings.Join(parts[1:], ":")
	}

	return imageName, "latest"
}
