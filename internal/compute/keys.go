package compute

import "fmt"

func corrKey(ns, payloadID string) string {
	return fmt.Sprintf("%s:corr:%s", ns, payloadID)
}

func systemKey(ns, systemID string) string {
	return fmt.Sprintf("%s:system:%s", ns, systemID)
}

func systemsSet(ns string) string {
	return fmt.Sprintf("%s:systems", ns)
}

func configEpochKey(ns string) string {
	return fmt.Sprintf("%s:control:config_epoch", ns)
}

func aliveKey(ns, instanceID string) string {
	return fmt.Sprintf("%s:alive:%s", ns, instanceID)
}

func statsKey(ns, instanceID string) string {
	return fmt.Sprintf("%s:stats:%s", ns, instanceID)
}

func aliveSet(ns string) string {
	return fmt.Sprintf("%s:alive", ns)
}

func statsSet(ns string) string {
	return fmt.Sprintf("%s:stats", ns)
}
