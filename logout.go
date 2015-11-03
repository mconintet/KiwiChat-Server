package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mconintet/kiwi"
)

func logoutUser(uid int64) error {
	db, err := sql.Open("mysql", dbConnInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// set token/network to NULL
	stmtUpe, err := db.Prepare("UPDATE user SET token=?,network=? WHERE id=?")
	if err != nil {
		return err
	}
	defer stmtUpe.Close()

	_, err = stmtUpe.Exec(nil, nil, uid)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot update token of user %d err: %s", uid, err.Error()))
	}

	return nil
}

func resetUserNetwork(uid int64) error {
	db, err := sql.Open("mysql", dbConnInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// set token/network to NULL
	stmtUpe, err := db.Prepare("UPDATE user SET network=? WHERE id=?")
	if err != nil {
		return err
	}
	defer stmtUpe.Close()

	_, err = stmtUpe.Exec(nil, uid)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot update network of user %d err: %s", uid, err.Error()))
	}

	return nil
}

func handleLogout(j *simplejson.Json, s kiwi.MessageSender, m *kiwi.Message) error {
	uid, err := j.Get("uid").Int64()
	if err != nil {
		return ErrDeformedRequest
	}

	token, err := j.Get("token").String()
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
		sendNeedsLogin("logout-return", s)
		return nil
	}

	// remove connId from uid-connId map
	uidConnIdMap.delUid(uid)

	// logout user
	logoutUser(uid)

	retMsg := simplejson.New()
	retMsg.Set("action", "logout-return")
	retMsg.Set("errCode", ErrNone)
	retMsg.Set("errMsg", ErrCodeMsg[ErrNone])

	byts, _ := retMsg.MarshalJSON()
	s.SendWholeBytes(byts, false)
	return nil
}
