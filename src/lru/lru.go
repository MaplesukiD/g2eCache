package lru

import "container/list"

// 使用map和双向链表来实现lru算法
type Cache struct {
	//允许使用的最大内存
	maxBytes int64
	//已使用内存
	nbytes int64
	//双向链表
	dl *list.List
	//字典,key是string,value是指针
	cache map[string]*list.Element
	//记录被移除的回调函数
	OnEvicted func(key string, value Value)
}

// New 初始化
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		dl:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add 增改
func (c *Cache) Add(key string, value Value) {
	//存在即修改,不在即新增
	if ele, ok := c.cache[key]; ok {
		c.dl.MoveToFront(ele)
		kv := ele.Value.(*node)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.dl.PushFront(&node{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// RemoveOldest 淘汰最近最少访问节点，即队末节点
func (c *Cache) RemoveOldest() {
	ele := c.dl.Back()
	if ele != nil {
		//从队列中移除
		c.dl.Remove(ele)
		kv := ele.Value.(*node)
		//从map中移除映射关系
		delete(c.cache, kv.key)
		//释放内存
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}

}

// Get 查,将查到的节点更新，放到队列队首
func (c *Cache) Get(key string) (value Value, ok bool) {
	ele, ok := c.cache[key]
	if ok {
		//移至队首
		c.dl.MoveToFront(ele)
		//取出value返回
		kv := ele.Value.(*node)
		return kv.value, true
	}
	return
}

// Len 获取数据个数
func (c *Cache) Len() int {
	return c.dl.Len()
}

// 双向链表节点
type node struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}
