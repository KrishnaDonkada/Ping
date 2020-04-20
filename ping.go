package main

import
(
	"flag"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"os/signal"
	"sync"
	"syscall"
   // "flag"
	"net"
	"os"
	"time"
	)

//Structure for reply packet
type reply struct {
	echo * icmp.Echo
	nunBytes int
	recvTime time.Time
	ttl int
}

//ping structure
type ping struct {
      hostAddr * net.IPAddr
      IPAddr string
      IP net.IP
      iptype int
      replychan chan * reply
	  wg sync.WaitGroup
      done chan bool
      done1 chan int
      done2 chan int
      done3 chan int
	  KVStore map[int]time.Time
      Seq int
      delay time.Duration
      packetssent int
      packetsrecived int
	  mux sync.Mutex
      TTL int
}

//function for sending the icmp packets
func (sender * ping) sendIcmp(con *icmp.PacketConn){
	defer sender.wg.Done()
	var add net.Addr
	add=sender.hostAddr
    for  {
		select {
		    case <-sender.done2:
		    	//fmt.Println("case done2")
				return
		default:
			echo := &icmp.Echo{1, sender.Seq, []byte("Hello! How do you do")}
			var message icmp.Message
			if sender.iptype == 4 {
				message.Type = ipv4.ICMPTypeEcho
				message.Body = echo
				message.Code = 0
			} else {
				message.Type = ipv6.ICMPTypeEchoRequest
				message.Body = echo
				message.Code = 0
			}
			msg, err := message.Marshal(nil)
			if err != nil {
				fmt.Println("Issue while marshalling the message:", err)
				sender.done<-false
				return
			}
			_, err = con.WriteTo(msg, add)
			//fmt.Println("destination",add)
			if err != nil {
				fmt.Println("Unable to ping the host:", err)
				sender.done<-false
				return
			}
			sent_time := time.Now()
			sender.mux.Lock()
			sender.KVStore[sender.Seq] = sent_time
			sender.mux.Unlock()
			sender.Seq++
			sender.packetssent++
			time.Sleep(sender.delay)
		}

	}
}

//function for recieving the icmp messages
func (sender * ping) recvIcmp(con *icmp.PacketConn){

	defer sender.wg.Done()
	for {
		select {
		case <-sender.done3:
			//fmt.Println("case done3")
			return

		default:
			readbuffer := make([]byte, 1024)
			err := con.SetReadDeadline(time.Now().Add(time.Second * 5))
			if err != nil {
				fmt.Println("Issue with setting read deadline", err)
				sender.done<-false
				return
			}
			var number int
			var readerr error
			var ttl int
			var cm *ipv4.ControlMessage
			var cm6 *ipv6.ControlMessage
			if sender.iptype == 4 {
				number,cm, _, readerr = con.IPv4PacketConn().ReadFrom(readbuffer)
				} else {
				number,cm6, _, readerr = con.IPv6PacketConn().ReadFrom(readbuffer)
			}
			//fmt.Println("here",readerr)
			if readerr != nil {
				error_op := readerr.(*net.OpError)
				if error_op.Timeout() {
					continue
				} else {
					//sender.done1 <- 1
					fmt.Println(readerr)
					sender.done<-false
					return
				}
			}

			recvtime := time.Now()
			var protocol int
			if sender.iptype == 4 {
				protocol = 1
				ttl=cm.TTL
			} else {
				protocol = 58
				ttl = cm6.HopLimit
			}
			parsed, error1 := icmp.ParseMessage(protocol, readbuffer)
			if error1 != nil {
				fmt.Println("Issue with Parsing the message error", error1)
				sender.done<-false
				return
			}
			if parsed.Type == ipv6.ICMPTypeEchoReply || parsed.Type == ipv4.ICMPTypeEchoReply {

				echobody := parsed.Body.(*icmp.Echo)
				sender.replychan <- &reply{&icmp.Echo{ID: echobody.ID, Seq: echobody.Seq, Data: echobody.Data}, number, recvtime,ttl}
			} else if parsed.Type== ipv4.ICMPTypeTimeExceeded {
				echobody := parsed.Body.(*icmp.TimeExceeded).Data
				if len(echobody) >= 28 {
					fmt.Println("Time exceeded for message:"," icmp_seq:",echobody[27])
				} else {
					fmt.Println("Time exceeded for message:")
				}

			} else if parsed.Type==ipv6.ICMPTypeTimeExceeded {
				echobody := parsed.Body.(*icmp.TimeExceeded).Data
				if len(echobody) >= 48 {
					fmt.Println("Time exceeded for message:"," icmp_seq:",echobody[47])
				} else {
					fmt.Println("Time exceeded for message:")
				}
			}
		}
	}
}

//main function
func main() {

	TTLvalue := flag.Int("ttl", 64, "Provide a Int TTL value ")
	Delay := flag.Duration("d",time.Second*1,"Provide the value for periodic delay between the packets")

	flag.Parse()
	//fmt.Println("ttvalue",*TTLvalue)
	ipvals := flag.Args()
	Inputaddress := ipvals[0]
	sender := &ping{}


	if len(os.Args) < 2 {
		fmt.Println("usage: ping.go <ip-address> <TTL_value>")
		return
	}
	sender.replychan = make(chan *reply, 10)
	sender.packetssent = 0
	sender.packetsrecived = 0
	sender.done1 = make(chan int,3)
	sender.done2 = make(chan int,3)
	sender.done3 = make(chan int,3)
	sender.done = make(chan bool,5)
	sender.TTL = *TTLvalue
	sender.KVStore = make(map[int]time.Time)
	hostAddr, err := net.ResolveIPAddr("ip", Inputaddress)
	if err != nil {
		fmt.Println("Issue with Resolving the ip address")
		return
	}
	var con *icmp.PacketConn
	sender.hostAddr = hostAddr
	sender.IPAddr = hostAddr.String()
	sender.IP = hostAddr.IP
	sender.Seq = 1
	sender.delay = *Delay
	sender.checkIpType()
	go sender.make(con)
	sender.wg.Add(1)
	sigs := make(chan os.Signal, 1)
	//done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		sender.done <- true
	}()
	x := <-sender.done
	//fmt.Println("received signa")
	//sender.done1<-1
	//sender.done2<-2
	//sender.done3<-3
	//fmt.Println("sent to the channels")

	//sender.wg.Wait()
	//fmt.Println("wait done")
	sender.closeConnection(con)
	if x {
		fmt.Println("-----------ping stats-----------------")
		fmt.Println("Packets sent:", sender.packetssent, "Packets Received:", sender.packetsrecived, "Number of lost packets:", sender.packetssent-sender.packetsrecived)
	}
}

//function to check if the ipv4 or ipv6
func (sender *ping) checkIpType() {
	if sender.IP.To4() != nil {
		sender.iptype=4
	}
}

//function to print the incoming packets
func (sender * ping) parseMessage() {
	//fmt.Println("Started ")
	defer sender.wg.Done()
	for {
		select {
		case packet:=<-sender.replychan:
            sender.packetsrecived++
			sender.mux.Lock()
            fmt.Println("Reply from",sender.IPAddr,": bytes:",packet.nunBytes," icmp_seq:",packet.echo.Seq," TTL:",packet.ttl," RTT:",packet.recvTime.Sub(sender.KVStore[packet.echo.Seq]))
            delete(sender.KVStore,packet.echo.Seq)
			sender.mux.Unlock()
        case <-sender.done1:
        	//fmt.Println("case done1")
			return
		default:
		     continue
		}
	}
}

//make function to drive the program
func (sender * ping) make(con * icmp.PacketConn) {
	defer sender.wg.Done()
	fmt.Println("Pinging ",sender.IPAddr,"with TTL value as:",sender.TTL," and delay as",sender.delay)
	if sender.iptype == 4 {
		var err error
		con,err = icmp.ListenPacket("ip4:1","")
		if err != nil {
			fmt.Println("Issue while establishing the listen connection",err)
			sender.done<-false
			//fmt.Println("Sent to the channel")
			return
		}

		err = con.IPv4PacketConn().SetControlMessage(ipv4.FlagTTL, true)
		if err != nil {
			fmt.Println("Issue with setting TTL Flag",err)
		}

		err =con.IPv4PacketConn().SetTTL(sender.TTL)
		if err != nil {
			fmt.Println("Issue with setting TTL Value",err)
		}
		err =con.IPv4PacketConn().SetMulticastTTL(sender.TTL)
		if err != nil {
			fmt.Println("Issue with setting TTL Value",err)
		}

	} else {
		var err error
		con,err = icmp.ListenPacket("ip6:58","")
		if con == nil {
			fmt.Println("Issue while establishing the listen connection",err)
			sender.done<-false
			return
		}

		err = con.IPv6PacketConn().SetControlMessage(ipv6.FlagHopLimit, true)
		if err != nil {
			fmt.Println("Issue with setting HopLimit Flag",err)
		}
		err= con.IPv6PacketConn().SetHopLimit(sender.TTL)
		if err != nil {
			fmt.Println("Issue with setting Hop limit Value",err)
		}
	}

	go sender.parseMessage()
	sender.wg.Add(1)
	go sender.recvIcmp(con)
	sender.wg.Add(1)
	go sender.sendIcmp(con)
	sender.wg.Add(1)
	//sender.wg.Wait()

}

//func to close the connection
func (sender * ping) closeConnection(con * icmp.PacketConn) {
	        //fmt.Println("Inside Close connection")
	    	err:=con.Close()
			con = nil
			if err != nil {
				//fmt.Println("Unable to close the connection:",err)
			}
}