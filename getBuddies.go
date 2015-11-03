package main

import (
	"database/sql"
	"github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mconintet/kiwi"
)

func handGetBuddies(j *simplejson.Json, s kiwi.MessageSender, m *kiwi.Message) error {
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
		sendNeedsLogin("get-buddies-return", s)
		return nil
	}

	// get buddies
	stmtOut, err := db.Prepare(`SELECT U.id, U.nickname, U.network, U.avatar
	FROM user AS U JOIN user_buddy as B ON U.id=B.buddy WHERE B.user=?`)
	if err != nil {
		return err
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(uid)
	if err != nil {
		return err
	}

	buddies := []*simplejson.Json{}
	for rows.Next() {
		var uid int
		var nickname sql.NullString
		var network sql.NullInt64
		var avatar sql.RawBytes

		err = rows.Scan(&uid, &nickname, &network, &avatar)

		if err != nil {
			return err
		}

		buddy := simplejson.New()
		buddy.Set("uid", uid)
		buddy.Set("nickname", nickname.String)
		buddy.Set("network", network.Int64)
		buddy.Set("avatar", encodeAvatar(avatar))
		buddies = append(buddies, buddy)
	}

	retMsg := simplejson.New()
	retMsg.Set("action", "get-buddies-return")
	retMsg.Set("buddies", buddies)
	retMsg.Set("errCode", ErrNone)
	retMsg.Set("errMsg", ErrCodeMsg[ErrNone])

	byts, _ := retMsg.MarshalJSON()
	s.SendWholeBytes(byts, false)
	return nil
}
