package main

import (
	"database/sql"
	"github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mconintet/kiwi"
)

func handRegister(j *simplejson.Json, s kiwi.MessageSender, m *kiwi.Message) error {
	email, err := j.Get("email").String()
	if err != nil {
		return ErrDeformedRequest
	}

	nickname, err := j.Get("nickname").String()
	if err != nil {
		return ErrDeformedRequest
	}

	password, err := j.Get("password").String()
	if err != nil {
		return ErrDeformedRequest
	}

	db, err := sql.Open("mysql", dbConnInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// validate login info
	stmtOut, err := db.Prepare("SELECT COUNT(*) FROM user WHERE email=?")
	if err != nil {
		return err
	}
	defer stmtOut.Close()

	var count int
	err = stmtOut.QueryRow(email).Scan(&count)
	if err != nil {
		return err
	}

	if count != 0 {
		retMsg := simplejson.New()
		retMsg.Set("action", "register-return")
		retMsg.Set("uid", -1)
		retMsg.Set("token", "")
		retMsg.Set("errCode", ErrEmailAlreadyExists)
		retMsg.Set("errMsg", ErrCodeMsg[ErrEmailAlreadyExists])

		byts, _ := retMsg.MarshalJSON()
		s.SendWholeBytes(byts, false)
		return nil
	}

	// make a default avatar with random color
	avatar := makeDefaultAvatar()

	stmtIns, err := db.Prepare("INSERT INTO user (email, password, nickname, token, avatar) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmtIns.Close()

	token := makeToken()
	res, err := stmtIns.Exec(email, makePassword(password), nickname, token, avatar)
	if err != nil {
		return err
	}

	uid, _ := res.LastInsertId()

	// associate new conn id with user
	uidConnIdMap.add(uid, s.GetConn().ID)

	retMsg := simplejson.New()
	retMsg.Set("action", "register-return")
	retMsg.Set("uid", uid)
	retMsg.Set("nickname", nickname)
	retMsg.Set("avatar", encodeAvatar(avatar))
	retMsg.Set("token", token)
	retMsg.Set("errCode", ErrNone)
	retMsg.Set("errMsg", ErrCodeMsg[ErrNone])

	byts, _ := retMsg.MarshalJSON()
	s.SendWholeBytes(byts, false)
	return nil
}
