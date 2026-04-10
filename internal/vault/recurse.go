package vault

import (
	"fmt"
	"strings"
)

// RecurseSecrets walks a Vault path recursively and returns all leaf secret paths.
// It uses ListSecrets to enumerate keys and recurses into directories (keys ending with "/").
func (c *Client) RecurseSecrets(mount, subpath string, engineType EngineType) ([]string, error) {
	keys, err := c.ListSecrets(mount, subpath, engineType)
	if err != nil {
		return nil, fmt.Errorf("listing %s/%s: %w", mount, subpath, err)
	}

	var paths []string
	for _, key := range keys {
		if strings.HasSuffix(key, "/") {
			// Directory — recurse
			childSubpath := strings.TrimSuffix(subpath+"/"+key, "/")
			childSubpath = strings.TrimPrefix(childSubpath, "/")
			children, err := c.RecurseSecrets(mount, childSubpath, engineType)
			if err != nil {
				return nil, err
			}
			paths = append(paths, children...)
		} else {
			// Leaf secret
			leaf := key
			if subpath != "" {
				leaf = subpath + "/" + key
			}
			paths = append(paths, leaf)
		}
	}
	return paths, nil
}
