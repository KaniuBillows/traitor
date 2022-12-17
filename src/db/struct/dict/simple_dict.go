package dict

type SimpleDict struct {
	m map[string]any
}

func MakeSimple() *SimpleDict {
	return &SimpleDict{
		m: make(map[string]any),
	}
}

func (d *SimpleDict) Get(key string) (val any, exists bool) {
	val, exists = d.m[key]
	return
}

func (d *SimpleDict) Len() int {
	return len(d.m)
}

func (d *SimpleDict) Put(key string, val any) (result int) {
	_, exists := d.m[key]
	d.m[key] = val
	if exists {
		return 0
	} else {
		return 1
	}
}

func (d *SimpleDict) PutIfAbsent(key string, val any) (result int) {
	_, exists := d.m[key]
	if exists {
		return 0
	}
	d.m[key] = val
	return 1
}

func (d *SimpleDict) PutIfExists(key string, val any) (result int) {
	_, exists := d.m[key]
	if exists == false {
		return 0
	}
	d.m[key] = val
	return 1
}

func (d *SimpleDict) Remove(key string) (result int) {
	_, exists := d.m[key]
	if exists {
		delete(d.m, key)
		return 1
	}
	return 0
}
func (d *SimpleDict) ForEach(consumer Consumer) {
	for k, v := range d.m {
		ctu := consumer(k, v)
		if ctu == false {
			return
		}
	}
}
func (d *SimpleDict) Keys() []string {
	var l = len(d.m)
	keys := make([]string, l)
	for k := range d.m {
		keys = append(keys, k)
	}
	return keys
}

func (d *SimpleDict) RandomKeys(limit int) []string {
	if limit >= d.Len() {
		return d.Keys()
	}
	keys := make([]string, limit)
	if limit <= 0 {
		return keys
	}
	i := 0
	for k := range d.m {
		keys[i] = k
		i++
		if limit == i {
			return keys
		}
	}
	return keys
}

func (d *SimpleDict) RandomDistinctKeys(limit int) []string {
	return d.RandomKeys(limit)
}
func (d *SimpleDict) Clear() {
	d.m = make(map[string]any)
}
