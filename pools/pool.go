package pools

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
	"errors"
)
var nowFunc = time.Now

type Pool struct {
	// 建立tcp连接
	Dial func() (net.Conn, error)

	// 健康检测，判断连接是否断开
	TestOnBorrow func(c net.Conn, t time.Time) error

	// 连接池中最大空闲连接数
	MaxIdle int

	// 打开最大的连接数
	MaxActive int

    // Idle多久断开连接，小于服务器超时时间
	IdleTimeout time.Duration

    // 配置最大连接数的时候，并且wait是true的时候，超过最大的连接，get的时候会阻塞，知道有连接放回到连接池
	Wait bool

	// 超过多久时间 链接关闭
	MaxConnLifetime time.Duration

	chInitialized uint32 // set to 1 when field ch is initialized 原子锁ch初始化一次

	mu     sync.Mutex    // 锁
	closed bool          // set to true when the pool is closed.
	Active int           // 连接池中打开的连接数
	ch     chan struct{} // limits open connections when p.Wait is true
	Idle   idleList      // idle 连接
}

// 空闲连，记录poolConn的头和尾
type idleList struct {
	count       int
	front, back *poolConn
}

// 连接的双向链表
type poolConn struct {
	C          net.Conn
	t          time.Time // idle 时间，即放会pool的时间
	created    time.Time //创建时间
	next, prev *poolConn
}


func (p *Pool) lazyInit() {

	if atomic.LoadUint32(&p.chInitialized) == 1 {
		return
	}
	p.mu.Lock()
	if p.chInitialized == 0 {
		p.ch = make(chan struct{}, p.MaxActive)
		if p.closed {
			close(p.ch)
		} else {
			for i := 0; i < p.MaxActive; i++ {
				p.ch <- struct{}{}
			}
		}
		atomic.StoreUint32(&p.chInitialized, 1)
	}
	p.mu.Unlock()
}

func (p *Pool) Get() (*poolConn, error) {

	// p.Wait == true. 的时候限制最大连接数
	if p.Wait && p.MaxActive > 0 {
		p.lazyInit()
		<-p.ch
	}

	p.mu.Lock()

	// 删除idle超时的连接，删除掉
	if p.IdleTimeout > 0 {
		n := p.Idle.count
		for i := 0; i < n && p.Idle.back != nil && p.Idle.back.t.Add(p.IdleTimeout).Before(nowFunc()); i++ {
			pc := p.Idle.back
			p.Idle.popBack()
			p.mu.Unlock()
			pc.C.Close()
			p.mu.Lock()
			p.Active--
		}
	}

	// Get Idle connection from the front of Idle list.
	for p.Idle.front != nil {
		pc := p.Idle.front
		p.Idle.popFront()
		p.mu.Unlock()
		if (p.TestOnBorrow == nil || p.TestOnBorrow(pc.C, pc.t) == nil) &&
			(p.MaxConnLifetime == 0 || nowFunc().Sub(pc.created) < p.MaxConnLifetime) {
			return pc, nil
		}
		pc.C.Close()
		p.mu.Lock()
		p.Active--
	}

	//pool关闭后直接return error
	if p.closed {
		p.mu.Unlock()
		return nil, errors.New("get on closed pool")
	}

	// Handle limit for p.Wait == false.
	if !p.Wait && p.MaxActive > 0 && p.Active >= p.MaxActive {
		p.mu.Unlock()
		return nil, errors.New("pool 耗尽了")
	}

	p.Active++
	p.mu.Unlock()
	c, err := p.Dial()
	if err != nil {
		c = nil
		p.mu.Lock()
		p.Active--
		if p.ch != nil && !p.closed {
			p.ch <- struct{}{}
		}
		p.mu.Unlock()
	}
	return &poolConn{C: c, created: nowFunc()}, err
}

//放回到连接池，如果强制关闭，不放回到连接池直接关闭连接
func (p *Pool) Put(pc *poolConn, forceClose bool) error {
	p.mu.Lock()
	if !p.closed && !forceClose {
		pc.t = nowFunc()
		p.Idle.pushFront(pc)
		if p.Idle.count > p.MaxIdle {
			pc = p.Idle.back
			p.Idle.popBack()
		} else {
			pc = nil
		}
	}

	if pc != nil {
		p.mu.Unlock()
		pc.C.Close()
		p.mu.Lock()
		p.Active--
	}

	if p.ch != nil && !p.closed {
		p.ch <- struct{}{}
	}
	p.mu.Unlock()
	return nil
}

// 关闭pool 只关闭空闲的连接，使用中的等放回到连接池的时候关闭
func (p *Pool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.Active -= p.Idle.count
	pc := p.Idle.front
	p.Idle.count = 0
	p.Idle.front, p.Idle.back = nil, nil
	if p.ch != nil {
		close(p.ch)
	}
	p.mu.Unlock()
	for ; pc != nil; pc = pc.next {
		pc.C.Close()
	}
	return nil
}

func (l *idleList) pushFront(pc *poolConn) {
	pc.next = l.front
	pc.prev = nil
	if l.count == 0 {
		l.back = pc
	} else {
		l.front.prev = pc
	}
	l.front = pc
	l.count++
	return
}

func (l *idleList) popFront() {
	pc := l.front
	l.count--
	if l.count == 0 {
		l.front, l.back = nil, nil
	} else {
		pc.next.prev = nil
		l.front = pc.next
	}
	pc.next, pc.prev = nil, nil
}

func (l *idleList) popBack() {
	pc := l.back
	l.count--
	if l.count == 0 {
		l.front, l.back = nil, nil
	} else {
		pc.prev.next = nil
		l.back = pc.prev
	}
	pc.next, pc.prev = nil, nil
}
