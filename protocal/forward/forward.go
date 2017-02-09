package forward

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/service-exposer/exposer"
)

const (
	CMD_AUTH          = "auth"
	CMD_AUTH_REPLY    = "auth:reply"
	CMD_FORWARD       = "forward"
	CMD_FORWARD_REPLY = "forward:reply"
)

type Reply struct {
	OK  bool
	Err string
}

type Auth struct {
	Key string
}

type Forward struct {
	Network string
	Address string
}

func ServerSide(authFn func(string) bool) exposer.HandshakeHandleFunc {
	return func(proto *exposer.Protocal, cmd string, details []byte) error {
		log.Print("server side cmd", cmd)
		switch cmd {
		case CMD_AUTH:
			var auth Auth
			err := json.Unmarshal(details, &auth)
			if err != nil {
				return err
			}

			if authFn(auth.Key) == true {
				return proto.Reply(CMD_AUTH_REPLY, &Reply{
					OK: true,
				})
			}

			err = proto.Reply(CMD_AUTH_REPLY, &Reply{
				OK:  false,
				Err: "auth failure",
			})
			if err != nil {
				return err
			}

			return errors.New("auth failure")
		case CMD_FORWARD:
			var forward Forward
			err := json.Unmarshal(details, &forward)
			if err != nil {
				return err
			}

			conn, err := net.Dial(forward.Network, forward.Address)
			if err != nil {
				proto.Reply(CMD_FORWARD_REPLY, &Reply{
					OK:  false,
					Err: err.Error(),
				})
				return err
			}
			conn.Close()

			err = proto.Reply(CMD_FORWARD_REPLY, &Reply{
				OK: true,
			})
			if err != nil {
				return err
			}

			/*
				ln := proto.Multiplex(false)
				defer ln.Close()

				for {
					local_conn, err := ln.Accept()
					if err != nil {
						log.Print("server session listener.Accept() err ", err)
						return err
					}
					defer local_conn.Close()

					log.Print("client new local conn ")

					remote_conn, err := net.Dial(forward.Network, forward.Address)
					if err != nil {
						log.Print("server remote session.Open() err ", err)
						return err
					}
					defer remote_conn.Close()

					go func() {
						wg := &sync.WaitGroup{}
						wg.Add(2)
						go func() {
							defer wg.Done()
							io.Copy(local_conn, remote_conn)
						}()

						go func() {
							defer wg.Done()
							io.Copy(remote_conn, local_conn)
						}()
						wg.Wait()
					}()
				}
			*/
			exposer.Serve(proto.Multiplex(false), func(conn net.Conn) exposer.ProtocalHandler {
				proto := exposer.NewProtocal(conn)
				proto.On = func(proto *exposer.Protocal, cmd string, details []byte) error {
					log.Print("server side :forward ", cmd)
					err := proto.Reply("", nil)
					if err != nil {
						return err
					}

					conn, err := net.Dial(forward.Network, forward.Address)
					if err != nil {
						return err
					}

					fmt.Println("forward to", forward.Address)
					proto.Forward(conn)
					return nil
				}
				return proto
			})

		}
		return errors.New("unknow cmd")
	}

}

func ClientSide(forward Forward, ln net.Listener) exposer.HandshakeHandleFunc {
	return func(proto *exposer.Protocal, cmd string, details []byte) error {
		log.Print("client side ", cmd)
		switch cmd {
		case CMD_AUTH_REPLY:
			var reply Reply
			err := json.Unmarshal(details, &reply)
			if err != nil {
				return err
			}

			if !reply.OK {
				return errors.New(reply.Err)
			}

			return proto.Reply(CMD_FORWARD, &forward)
		case CMD_FORWARD_REPLY:

			var reply Reply
			err := json.Unmarshal(details, &reply)
			if err != nil {
				return err
			}

			if !reply.OK {
				return errors.New(reply.Err)
			}

			session := proto.Multiplex(true)

			for {
				local_conn, err := ln.Accept()
				if err != nil {
					log.Print("client local listener.Accept() err ", err)
					return err
				}

				log.Print("client new local conn ", local_conn)

				remote_conn, err := session.Open()
				if err != nil {
					log.Print("client remote session.Open() err ", err)
					return err
				}

				log.Print("client new session conn")

				/*
					go func() {
						wg := &sync.WaitGroup{}
						wg.Add(2)
						go func() {
							defer wg.Done()
							io.Copy(local_conn, io.TeeReader(remote_conn, os.Stdout))
						}()

						go func() {
							defer wg.Done()
							io.Copy(remote_conn, local_conn)
						}()
						wg.Wait()
					}()
				*/

				proto_forward := exposer.NewProtocal(remote_conn)
				proto_forward.On = func(proto *exposer.Protocal, cmd string, details []byte) error {
					log.Println("client side :forward ", cmd)
					log.Println("client new conn")
					proto.Forward(local_conn)
					return nil
				}

				go proto_forward.Request("", nil)
			}
		default:
			return errors.New("unknow cmd")
		}
	}

}