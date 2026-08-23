package pgpool

import (
	"context"
	"fmt"
	"time"
)



func (p *Pool) connectionJanitor() {
	ticker := time.NewTicker(time.Second*2)
	defer ticker.Stop()
	for range ticker.C {
		p.mutex.Lock()
		if p.isClosing {
			p.mutex.Unlock()
			return
		}
		// if debugMode {
		// 	fmt.Printf("Connection janitor work\n")
		// }
		// mp.mutex.Lock()
		cjs := make([]*ConnResource, len(p.expiredConnections))
		copy(cjs, p.expiredConnections)
		p.expiredConnections = []*ConnResource{}
		p.mutex.Unlock()
		fmt.Println(p.expiredConnections)
		fmt.Println(cjs, len(cjs))
		// в отдельной горутине закрываем соединения 
		if len(cjs) != 0 {
			go func (connJanitorSlice []*ConnResource)  {
				for _, ec := range connJanitorSlice {
					fmt.Println("Закрытие соединения", ec)
					ec.conn.Close(context.Background())
					// if debugMode {
					// 	mp.mutex.Lock()
					// 	debugClosedConnections++
					// 	mp.mutex.Unlock()
					// }
				}
			}(cjs)
		}
	}
}