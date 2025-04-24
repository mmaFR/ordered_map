package ordered_map

import (
	"encoding/json"
	"math"
	"sort"
	"sync"
)

type orderedStringMapEntry struct {
	value interface{}
	index uint64
}

func (osm *orderedStringMapEntry) MarshalJSON() ([]byte, error) {
	jb, err := json.Marshal(osm.value)
	return jb, err
}

type OrderedStringMap struct {
	data      map[string]*orderedStringMapEntry
	lastIndex uint64
	lock      sync.RWMutex
}

func (o *OrderedStringMap) MarshalJSON() ([]byte, error) {
	jb, err := json.Marshal(o.data)
	return jb, err
}

func (o *OrderedStringMap) Reindex() {
	o.lock.Lock()
	defer o.lock.Unlock()

	var keys []string = o.getKeys(false)
	for _, k := range keys {
		o.data[k].index = o.getIndex()
	}
}

func (o *OrderedStringMap) getIndex() uint64 {
	if o.lastIndex == math.MaxUint64 {
		o.lastIndex = 0
		o.Reindex()
	}
	o.lastIndex++
	return o.lastIndex
}

func (o *OrderedStringMap) Len() int {
	return len(o.data)
}

func (o *OrderedStringMap) Add(key string, value interface{}) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if v, e := o.data[key]; e {
		v.value = value
		v.index = o.getIndex()
	} else {
		o.data[key] = &orderedStringMapEntry{
			value: value,
			index: o.getIndex(),
		}
	}
}

func (o *OrderedStringMap) Del(key string) {
	o.lock.Lock()
	defer o.lock.Unlock()
	delete(o.data, key)
}

func (o *OrderedStringMap) Get(key string) (v interface{}, e bool) {
	o.lock.RLock()
	defer o.lock.RUnlock()
	_, e = o.data[key]
	if e {
		v = o.data[key].value
	}
	return
}

func (o *OrderedStringMap) GetKeys() []string {
	return o.getKeys(true)
}

func (o *OrderedStringMap) getKeys(locking bool) []string {
	if locking {
		o.lock.RLock()
		defer o.lock.RUnlock()
	}
	type item struct {
		key   string
		value *orderedStringMapEntry
	}
	var items []*item = make([]*item, len(o.data))
	var cursor uint64
	var sliceToReturn []string = make([]string, len(o.data))

	for k, v := range o.data {
		items[cursor] = &item{
			key:   k,
			value: v,
		}
		cursor++
	}

	sort.Slice(items, func(i, j int) bool { return items[i].value.index < items[j].value.index })

	for i, v := range items {
		sliceToReturn[i] = v.key
	}
	return sliceToReturn
}

func NewOrderedStringMap(cap ...int) *OrderedStringMap {
	if len(cap) > 0 {
		return &OrderedStringMap{
			data: make(map[string]*orderedStringMapEntry, cap[0]),
		}
	} else {
		return &OrderedStringMap{
			data: make(map[string]*orderedStringMapEntry),
		}
	}
}
