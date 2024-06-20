package singleflight

import "sync"

type call struct {
	waitGroup sync.WaitGroup
	val       any
	err       error
}

type Group struct {
	mutex sync.Mutex
	m     map[string]*call
}

// Do 实现针对相同的 key ，无论 Do 被调用多少次， fn 只被调用一次
func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mutex.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mutex.Unlock()
		//c.waitGroup不为0，即请求在进行中，则等待
		c.waitGroup.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.waitGroup.Add(1)
	g.m[key] = c
	g.mutex.Unlock()

	c.val, c.err = fn()
	c.waitGroup.Done()

	g.mutex.Lock()
	delete(g.m, key)
	g.mutex.Unlock()

	return c.val, c.err
}
