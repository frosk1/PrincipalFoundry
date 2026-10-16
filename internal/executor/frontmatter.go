package executor

import "strings"

// parseFrontmatter extracts simple "key: value" pairs from a leading YAML-style
// frontmatter block. It is intentionally minimal: it understands the flat
// scalar fields used by the factory control files and ignores lists, comments,
// and nested structures. Values are returned with surrounding quotes removed.
func parseFrontmatter(content string) map[string]string {
	fields := map[string]string{}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(strings.TrimSuffix(lines[0], "\r")) != "---" {
		return fields
	}

	for _, raw := range lines[1:] {
		line := strings.TrimSuffix(raw, "\r")
		if strings.TrimSpace(line) == "---" {
			break
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "-") {
			continue
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		fields[key] = value
	}

	return fields
}
