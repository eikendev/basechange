// Extracted from https://containrrr.dev/watchtower/

package registry

import (
	"fmt"
	url2 "net/url"
	"strings"
)

// convertToHostname strips a url from everything but the hostname part
func convertToHostname(url string) (string, string, error) {
	urlWithSchema := fmt.Sprintf("x://%s", url)
	u, err := url2.Parse(urlWithSchema)
	if err != nil {
		return "", "", err
	}
	hostName := u.Hostname()
	port := u.Port()

	return hostName, port, err
}

// NormalizeRegistry makes sure variations of DockerHubs registry
func NormalizeRegistry(registry string) (string, error) {
	hostName, port, err := convertToHostname(registry)
	if err != nil {
		return "", err
	}

	if hostName == "registry-1.docker.io" || hostName == "docker.io" {
		hostName = "index.docker.io"
	}

	if port != "" {
		return fmt.Sprintf("%s:%s", hostName, port), nil
	}
	return hostName, nil
}

// GetScopeFromImageName normalizes an image name for use as scope during auth and head requests
func GetScopeFromImageName(image, svc string) string {
	parts := strings.Split(image, "/")

	if len(parts) > 2 {
		if strings.Contains(svc, "docker.io") {
			return fmt.Sprintf("%s/%s", parts[1], strings.Join(parts[2:], "/"))
		}
		return strings.Join(parts, "/")
	}

	if len(parts) == 2 {
		if strings.Contains(parts[0], "docker.io") {
			return fmt.Sprintf("library/%s", parts[1])
		}
		return strings.Replace(image, svc+"/", "", 1)
	}

	if strings.Contains(svc, "docker.io") {
		return fmt.Sprintf("library/%s", parts[0])
	}

	return image
}
