package main

import (
	"database/sql"
	"github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mconintet/kiwi"
)

func handAddBuddy(j *simplejson.Json, s kiwi.MessageSender, m *kiwi.Message) error {
	uid, err := j.Get("uid").Int()
	if err != nil {
		return ErrDeformedRequest
	}

	token, err := j.Get("token").String()
	if err != nil {
		return ErrDeformedRequest
	}

	buddyID, err := j.Get("buddyID").Int()
	if err != nil {
		return ErrDeformedRequest
	}

	db, err := sql.Open("mysql", dbConnInfo)
	if err != nil {
		return err
	}
	defer db.Close()

	// validate request info
	stmtOut, err := db.Prepare("SELECT COUNT(*) FROM user WHERE id=? AND token=?")
	if err != nil {
		return err
	}

	var count int
	err = stmtOut.QueryRow(uid, token).Scan(&count)
	if err != nil {
		return err
	}
	stmtOut.Close()

	if count == 0 {
		return ErrIllegalRequest
	}

	// check if buddy is exists or not
	stmtOut, err = db.Prepare("SELECT COUNT(*) FROM user WHERE id=?")
	if err != nil {
		return err
	}
	defer stmtOut.Close()

	err = stmtOut.QueryRow(buddyID).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		retMsg := simplejson.New()
		retMsg.Set("action", "add-buddy-return")
		retMsg.Set("errCode", ErrBuddyNotFound)
		retMsg.Set("errMsg", ErrCodeMsg[ErrBuddyNotFound])

		byts, _ := retMsg.MarshalJSON()
		s.SendWholeBytes(byts, false)
		return nil
	}

	// check if relation is exists or not
	stmtOut, err = db.Prepare("SELECT COUNT(*) FROM user_buddy WHERE user=? AND buddy=?")
	if err != nil {
		return err
	}
	defer stmtOut.Close()

	err = stmtOut.QueryRow(uid, buddyID).Scan(&count)
	if err != nil {
		return err
	}

	if count == 1 {
		retMsg := simplejson.New()
		retMsg.Set("action", "add-buddy-return")
		retMsg.Set("errCode", ErrBuddyAlreadyExists)
		retMsg.Set("errMsg", ErrCodeMsg[ErrBuddyAlreadyExists])

		byts, _ := retMsg.MarshalJSON()
		s.SendWholeBytes(byts, false)
		return nil
	}

	stmtIns, err := db.Prepare("INSERT INTO user_buddy (user, buddy) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmtIns.Close()

	_, err = stmtIns.Exec(uid, buddyID)
	if err != nil {
		return err
	}

	// become friends each other
	_, err = stmtIns.Exec(buddyID, uid)
	if err != nil {
		return err
	}

	// return buddy info
	stmtOut, err = db.Prepare("SELECT nickname,network,avatar FROM user WHERE id=?")
	if err != nil {
		return err
	}

	var nickname sql.NullString
	var network sql.NullInt64
	var avatar []byte
	err = stmtOut.QueryRow(buddyID).Scan(&nickname, &network, &avatar)
	if err != nil {
		return err
	}

	retMsg := simplejson.New()
	retMsg.Set("action", "add-buddy-return")

	retMsg.SetPath([]string{"buddy", "uid"}, buddyID)
	retMsg.SetPath([]string{"buddy", "nickname"}, nickname.String)
	retMsg.SetPath([]string{"buddy", "network"}, network.Int64)
	retMsg.SetPath([]string{"buddy", "avatar"}, encodeAvatar(avatar))

	retMsg.Set("errCode", ErrNone)
	retMsg.Set("errMsg", ErrCodeMsg[ErrNone])

	byts, _ := retMsg.MarshalJSON()
	s.SendWholeBytes(byts, false)
	return nil
}
