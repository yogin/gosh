package service

// InstanceSort implements sort.Interface
type InstanceSort []*Instance

func (a InstanceSort) Len() int {
	return len(a)
}

func (a InstanceSort) Less(i, j int) bool {
	for _, key := range DefaultTags {
		if a[i].Tags[key] < a[j].Tags[key] {
			return true
		}

		if a[i].Tags[key] > a[j].Tags[key] {
			return false
		}
	}

	return a[i].ID < a[j].ID
}

func (a InstanceSort) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
