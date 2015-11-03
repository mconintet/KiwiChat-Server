package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mconintet/kiwi"
)

func handleLogin(j *simplejson.Json, s kiwi.MessageSender, m *kiwi.Message) error {
	email, err := j.Get("email").String()
	if err != nil {
		return ErrDeformedRequest
	}

	password, err := j.Get("password").String()
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

	// validate login info
	stmtOut, err := db.Prepare("SELECT id,nickname,avatar FROM user WHERE email=? AND password=?")
	if err != nil {
		return err
	}
	defer stmtOut.Close()

	var uid int64
	var nickname sql.NullString
	var avatar []byte
	err = stmtOut.QueryRow(email, makePassword(password)).Scan(&uid, &nickname, &avatar)
	if err != nil {
		if err == sql.ErrNoRows {
			retMsg := simplejson.New()
			retMsg.Set("action", "login-return")
			retMsg.Set("uid", -1)
			retMsg.Set("token", "")
			retMsg.Set("errCode", ErrInvalidLoginInfo)
			retMsg.Set("errMsg", ErrCodeMsg[ErrInvalidLoginInfo])

			byts, _ := retMsg.MarshalJSON()
			s.SendWholeBytes(byts, false)
			return nil
		}
		return err
	}

	// associate new conn id with user
	uidConnIdMap.add(uid, s.GetConn().ID)

	// update token/network
	token := makeToken()
	stmtUpe, err := db.Prepare("UPDATE user SET token=?,network=? WHERE id=?")
	if err != nil {
		return err
	}
	defer stmtUpe.Close()

	_, err = stmtUpe.Exec(token, network, uid)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot update token of user %d err: %s", uid, err.Error()))
	}

	retMsg := simplejson.New()
	retMsg.Set("action", "login-return")
	retMsg.Set("uid", uid)
	retMsg.Set("token", token)
	retMsg.Set("nickname", nickname.String)
	retMsg.Set("avatar", encodeAvatar(avatar))
	retMsg.Set("errCode", ErrNone)
	retMsg.Set("errMsg", ErrCodeMsg[ErrNone])

	byts, _ := retMsg.MarshalJSON()
	s.SendWholeBytes(byts, false)
	return nil
}
