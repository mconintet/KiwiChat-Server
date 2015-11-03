package main

import (
	"database/sql"
	"github.com/bitly/go-simplejson"
	"github.com/mconintet/kiwi"
)

func handleGetUserInfo(j *simplejson.Json, s kiwi.MessageSender, m *kiwi.Message) error {
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
		sendNeedsLogin("get-user-info-return", s)
		return nil
	}

	stmtOut, err := db.Prepare("SELECT nickname,avatar FROM user WHERE id=?")
	if err != nil {
		return err
	}
	defer stmtOut.Close()

	var nickname sql.NullString
	var avatar []byte
	err = stmtOut.QueryRow(uid).Scan(&nickname, &avatar)
	if err != nil {
		if err == sql.ErrNoRows {
			retMsg := simplejson.New()
			retMsg.Set("action", "get-user-info-return")
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

	retMsg := simplejson.New()
	retMsg.Set("action", "get-user-info-return")
	retMsg.Set("uid", uid)
	retMsg.Set("nickname", nickname.String)
	retMsg.Set("avatar", encodeAvatar(avatar))
	retMsg.Set("errCode", ErrNone)
	retMsg.Set("errMsg", ErrCodeMsg[ErrNone])

	byts, _ := retMsg.MarshalJSON()
	s.SendWholeBytes(byts, false)
	return nil
}
