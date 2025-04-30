package config

func mapKeys(m map[string]func(string) (interface{}, error)) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
