package pgpool

import (
	"context"
	"fmt"
	"time"

	"github.com/TiJon8/cpol/pgconn"
)


type ConnResource struct {
	conn *pgconn.PgConn
	lastUsed time.Time
}


func (p *Pool) Get(ctx context.Context) (c *ConnResource, err error) {
	p.mutex.Lock()
	if p.isClosing {
		return nil, fmt.Errorf("Pool is in closing state")
	}

	for len(p.store) > 0 {
		// проверить на протухшие соединения
		lastIdx := len(p.store)-1
		pConn := p.store[lastIdx]
		p.store = p.store[:lastIdx]

		// сюда попадут мертвые
		if time.Since(pConn.lastUsed) > p.maxConnIdleTimeout {
			// if debugMode {
			// 	fmt.Println("Соединение протухло")
			// 	debugIdleTimeoutConnections++
			// }
			p.expiredConnections = append(p.expiredConnections, pConn)
			fmt.Println("Всего соединений в сборщике", p.expiredConnections)
			p.oppenedConnections--
			continue
		}
		// тут значит есть кто-то живой
		p.wgFlight.Add(1)
		p.mutex.Unlock()
		// в принципе можно передать это какой нибудь горутине сборщику мусора условно, чтобы он в фоне закрывал сам соединения
		// for _, ec := range expiredConnections {
		// 	ec.conn.Close()
		// }

		return pConn, nil
	}

	if p.oppenedConnections < p.maxConns {
		p.oppenedConnections += 1
		// cохраняю услоынй номер коннекта, чтобы отпустить мьютекс
		// connNumber := p.oppenedConnections
		p.mutex.Unlock()

		pConn, err := pgconn.ConnectConfig(ctx, p.config.ConnectionConfig)
		if err != nil {
			return nil, err
		}
		p.wgFlight.Add(1)
		// if debugMode {
		// 	fmt.Println("Создание нового соединения")
		// 	p.mutex.Lock()
		// 	debugTotalConnections++
		// 	p.mutex.Unlock()
		// }
		return &ConnResource{
			conn: pConn,
			lastUsed: time.Now(),
		}, nil
	}
	bridge := make(chan *ConnResource, 1)
	t := p.nextT
	p.nextT++
	p.waitQueue[t] = bridge
	p.mutex.Unlock()

	select {
	case <-ctx.Done():
		p.mutex.Lock()
		delete(p.waitQueue, t)
		p.mutex.Unlock()
		select {
		case pConn, oppened := <-bridge:
			if !oppened {
				return nil, fmt.Errorf("Pool is in closing state")
			}
			p.Put(pConn) 
		default:
		}
		return nil, fmt.Errorf("Context has expired")
	case pConn, oppened := <- bridge:
		if !oppened {
			return nil, fmt.Errorf("Pool is in closing state")
		}
		return pConn, nil
	}
}


func (p *Pool) Put(c *ConnResource) error {
	c.lastUsed = time.Now()
	p.mutex.Lock()

	var ot uint64
	for p.ownerT < p.nextT {
		nextG, exists := p.waitQueue[p.ownerT]
		ot = p.ownerT
		p.ownerT++
		if exists {
			delete(p.waitQueue, ot)
			p.mutex.Unlock()
			nextG<-c
			return nil
		}
	}
	p.wgFlight.Done()
	p.store = append(p.store, c)
	p.mutex.Unlock()
	return nil
}
