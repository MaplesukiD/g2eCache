package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	//使用的hash算法
	hash Hash
	//虚拟节点倍数
	replicas int
	//哈希环
	keys []int
	//虚拟节点和真实节点映射表 key:虚拟节点hash值 value:真实节点名称
	hashMap map[int]string
}

// New 指定虚拟节点个数以及hash算法
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}

	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 添加真实节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			//生成虚拟节点，得到其hash值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			//在环上添加新节点
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get 根据虚拟节点key查找返回真实节点value
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
