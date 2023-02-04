// Extracted from https://containrrr.dev/watchtower/

package registry

import (
	"net/http"
)

// ContentDigestHeader is the key for the key-value pair containing the digest header
const ContentDigestHeader = "Docker-Content-Digest"

func GetImageDigest(image string) (string, error) {
	var digest string

	token, err := GetToken(image)
	if err != nil {
		return "", err
	}

	digestURL, err := BuildManifestURL(image)
	if err != nil {
		return "", err
	}

	if digest, err = getDigest(digestURL, token); err != nil {
		return "", err
	}

	return digest, nil
}

// GetDigest from registry using a HEAD request to prevent rate limiting
func getDigest(url string, token string) (string, error) {
	header := http.Header{"Authorization": []string{token}}
	header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v2+json")
	header.Add("Accept", "application/vnd.docker.distribution.manifest.v1+json")
	header.Add("Accept", "application/vnd.oci.image.index.v1+json")

	resp, err := RetryReq("HEAD", url, maxRetries+1, header, http.StatusOK)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return resp.Header.Get(ContentDigestHeader), nil
}
