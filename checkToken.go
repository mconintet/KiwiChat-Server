package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mconintet/kiwi"
)

func sendNeedsLogin(action string, s kiwi.MessageSender) {
	retMsg := simplejson.New()
	retMsg.Set("action", action)
	retMsg.Set("errCode", ErrInvalidToken)
	retMsg.Set("errMsg", ErrCodeMsg[ErrInvalidToken])

	byts, _ := retMsg.MarshalJSON()
	s.SendWholeBytes(byts, false)
}

func checkToken(uid int64, token string) (isValid bool, err error) {
	if token == "" {
		return false, nil
	}

	db, err := sql.Open("mysql", dbConnInfo)
	if err != nil {
		return false, err
	}
	defer db.Close()

	// valid user and his token
	stmtOut, err := db.Prepare("SELECT COUNT(*) FROM user WHERE id=? AND token=? AND token IS NOT NULL")
	if err != nil {
		return false, err
	}
	defer stmtOut.Close()

	var count int
	err = stmtOut.QueryRow(uid, token).Scan(&count)
	if err != nil {
		return false, err
	}

	if count != 1 {
		return false, nil
	}

	return true, nil
}

func handleCheckToken(j *simplejson.Json, s kiwi.MessageSender, m *kiwi.Message) error {
	uid, err := j.Get("uid").Int64()
	if err != nil {
		return ErrDeformedRequest
	}

	token, err := j.Get("token").String()
	if err != nil {
		return ErrDeformedRequest
	}

	network, err := j.Get("network").Int()
	if err != nil {
		return ErrDeformedRequest
	}

	db, err := sql.Open("mysql", dbConnInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// valid user and his token
	tokenIsValid, err := checkToken(uid, token)
	if err != nil {
		return err
	}

	if !tokenIsValid {
		sendNeedsLogin("check-token-return", s)
		return nil
	}

	// associate new conn id with user
	uidConnIdMap.add(uid, s.GetConn().ID)

	// update network
	stmtUpe, err := db.Prepare("UPDATE user SET network=? WHERE id=?")
	if err != nil {
		return err
	}
	defer stmtUpe.Close()

	_, err = stmtUpe.Exec(network, uid)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot update conn id of user %d err: %s", uid, err.Error()))
	}

	retMsg := simplejson.New()
	retMsg.Set("action", "check-token-return")
	retMsg.Set("errCode", ErrNone)
	retMsg.Set("errMsg", ErrCodeMsg[ErrNone])

	byts, _ := retMsg.MarshalJSON()
	s.SendWholeBytes(byts, false)
	return nil
}
