package layouts

type AssetOptions struct {
	CSS []string
	JS  []string
}

func CSS(paths ...string) AssetOptions {
	return AssetOptions{CSS: paths}
}

func JS(paths ...string) AssetOptions {
	return AssetOptions{JS: paths}
}

func buildAssetOptions(options ...AssetOptions) AssetOptions {
	merged := AssetOptions{}

	for _, option := range options {
		merged.CSS = append(merged.CSS, option.CSS...)
		merged.JS = append(merged.JS, option.JS...)
	}

	merged.CSS = uniqueNonEmpty(merged.CSS)
	merged.JS = uniqueNonEmpty(merged.JS)

	return merged
}

func uniqueNonEmpty(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	unique := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		unique = append(unique, path)
	}
	return unique
}
