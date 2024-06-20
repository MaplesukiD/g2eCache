package geecache

//                          是
//接收 key --> 检查是否被缓存 ----> 返回缓存值 ⑴
//                |  否                       是
//                |-----> 是否应当从远程节点获取 --> 使用一致性哈希选择节点      是                                    是
//                             |                        |---> 是否是远程节点 ---> HTTP 客户端访问远程节点 --> 成功？-----> 返回缓存值（2）
//                             |                                  | 否                                     | 否
//                             |                                  | <-------------------------------------|
//                             |  否                              ↓
//                             |-----> 本地调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶

import (
	"fmt"
	"goCache/src/geecache/singleflight"
	"log"
	"sync"
)

// 回调Getter
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
	loader    *singleflight.Group
}

var (
	mutex  sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mutex.Lock()
	defer mutex.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mutex.RLock()
	g := groups[name]
	mutex.RUnlock()
	return g
}

// Get 从mainCache中查找缓存，存在即返回，不存在调用load方法
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] 命中")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	//保证并发状态下，针对相同的key，只调用一次load
	view, err := g.loader.Do(key, func() (any, error) {
		// 使用 PickPeer方法选择节点，若非本机节点，则调用 getFromPeer从远程获取。
		if g.peers != nil {
			if peer, ok := g.peers.PeerPicker(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		// 本机节点，调用getLocally
		return g.getLocally(key)
	})
	if err == nil {
		return view.(ByteView), nil
	}
	return
}

// 调用回调函数获取源数据，将源数据添加在缓存中
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// 实现了 PeerPicker 接口的 HTTPPool 注入到 Group
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// 使用实现了 PeerGetter 接口的 httpGetter 从远程节点获取缓存值
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
