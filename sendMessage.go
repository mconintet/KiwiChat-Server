package main

import (
	"github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mconintet/kiwi"
)

func sendUnreachable(msgID int, s kiwi.MessageSender) error {
	retMsg := simplejson.New()
	retMsg.Set("action", "send-message-return")
	retMsg.Set("msgID", msgID)
	retMsg.Set("errCode", ErrCodeUnreachable)
	retMsg.Set("errMsg", ErrCodeMsg[ErrCodeUnreachable])

	byts, _ := retMsg.MarshalJSON()
	_, err := s.SendWholeBytes(byts, false)

	return err
}

func handleSendMessage(j *simplejson.Json, s kiwi.MessageSender, m *kiwi.Message) error {
	token, err := j.Get("token").String()
	if err != nil {
		return ErrDeformedRequest
	}

	from, err := j.Get("from").Int64()
	if err != nil {
		return ErrDeformedRequest
	}

	to, err := j.Get("to").Int64()
	if err != nil {
		return ErrDeformedRequest
	}

	msgID, err := j.Get("msgID").Int()
	if err != nil {
		return ErrDeformedRequest
	}

	msgType, err := j.Get("msgType").Int()
	if err != nil {
		return ErrDeformedRequest
	}

	msgContent, err := j.Get("msgContent").String()
	if err != nil {
		return ErrDeformedRequest
	}

	// valid user and his token
	tokenIsValid, err := checkToken(from, token)
	if err != nil {
		return err
	}

	if !tokenIsValid {
		sendNeedsLogin("send-message-return", s)
		return nil
	}

	// retrieve buddy's conn id
	buddyConnId, ok := uidConnIdMap.getConnId(to)
	if !ok {
		sendUnreachable(msgID, s)
		return nil
	}

	conn, ok := s.GetConn().Server.ConnPool.Get(buddyConnId)
	if !ok {
		sendUnreachable(msgID, s)
		return nil
	}

	msg := simplejson.New()
	msg.Set("action", "new-message")
	msg.Set("from", from)
	msg.Set("msgType", msgType)
	msg.Set("msgContent", msgContent)
	msg.Set("errCode", ErrNone)
	msg.Set("errMsg", ErrCodeMsg[ErrNone])

	sender := &kiwi.DefaultMessageSender{}
	sender.SetConn(conn)
	byts, _ := msg.MarshalJSON()
	_, err = sender.SendWholeBytes(byts, false)

	if err != nil {
		sendUnreachable(msgID, s)
	}
	return nil
}
