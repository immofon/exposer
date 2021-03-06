package auth

import (
	"encoding/json"
	"errors"

	"github.com/service-exposer/exposer"
	"github.com/service-exposer/exposer/protocal/route"
	"github.com/service-exposer/exposer/service"
)

const (
	CMD_AUTH       = "auth"
	CMD_AUTH_REPLY = "auth:reply"
)

type Reply struct {
	OK  bool
	Err string
}

type AuthReq struct {
	Key string
}

func ServerSide(router *service.Router, authFn func(key string) (allow bool)) exposer.HandshakeHandleFunc {
	return func(proto *exposer.Protocal, cmd string, details []byte) error {
		switch cmd {
		case CMD_AUTH:
			var req AuthReq
			err := json.Unmarshal(details, &req)
			if err != nil {
				return err
			}

			if !authFn(req.Key) {
				err := errors.New("auth: forbidden key")
				proto.Reply(CMD_AUTH_REPLY, &Reply{
					OK:  false,
					Err: err.Error(),
				})

				return err
			}

			err = proto.Reply(CMD_AUTH_REPLY, &Reply{
				OK: true,
			})
			if err != nil {
				return err
			}

			session := proto.Multiplex(false)
			for {
				conn, err := session.Accept()
				if err != nil {
					return err
				}

				proto_next := exposer.NewProtocalWithParent(proto, conn)
				proto_next.On = route.ServerSide(router)
				go proto_next.Handle()
			}

			return nil
		}

		return errors.New("unknow cmd: " + cmd)
	}
}

type NextRoute struct {
	Req        route.RouteReq
	HandleFunc exposer.HandshakeHandleFunc
	Cmd        string
	Details    interface{}
}

func ClientSide(routes <-chan NextRoute) exposer.HandshakeHandleFunc {
	return func(proto *exposer.Protocal, cmd string, details []byte) error {
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

			session := proto.Multiplex(true)

			for nr := range routes {
				conn, err := session.Open()
				if err != nil {
					return err
				}

				r := nr
				proto_next := exposer.NewProtocalWithParent(proto, conn)
				proto_next.On = route.ClientSide(r.HandleFunc, r.Cmd, r.Details)
				go proto_next.Request(route.CMD_ROUTE, &r.Req)
			}

			return nil
		}

		return errors.New("unknow cmd: " + cmd)
	}
}
