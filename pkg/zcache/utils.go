package zcache

import (
	"fmt"
	"strings"
)

const KeySplitter = "/"

func getKeyWithPrefix(prefix, key string) string {
	if prefix != "" {
		return fmt.Sprintf("%s%s%s", prefix, KeySplitter, key)
	}
	return key
}

func getKeysWithPrefix(prefix string, keys []string) []string {
	if prefix != "" {
		var newKeys []string
		for _, key := range keys {
			newKeys = append(newKeys, fmt.Sprintf("%s%s%s", prefix, KeySplitter, key))
		}
		return newKeys
	}

	return keys
}

func stripPrefixFromKeys(prefix string, keys []string) []string {
	if prefix == "" {
		return keys
	}
	fullPrefix := prefix + KeySplitter
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, strings.TrimPrefix(key, fullPrefix))
	}
	return result
}
