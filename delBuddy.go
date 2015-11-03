package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mconintet/kiwi"
)

func handDelBuddy(j *simplejson.Json, s kiwi.MessageSender, m *kiwi.Message) error {
	uid, err := j.Get("uid").Int()
	if err != nil {
		return ErrDeformedRequest
	}

	token, err := j.Get("token").String()
	if err != nil {
		return ErrDeformedRequest
	}

	buddyID, err := j.Get("buddyID").String()
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
	defer stmtOut.Close()

	var count int
	err = stmtOut.QueryRow(uid, token).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrIllegalRequest
	}

	// delete relation
	stmtUpe, err := db.Prepare("DELETE FROM user_buddy WHERE user=? AND buddy=?")
	if err != nil {
		return err
	}
	defer stmtUpe.Close()

	_, err = stmtUpe.Exec(uid, buddyID)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot delete relation user %d buddy:%d err: %s", uid, buddyID, err.Error()))
	}

	retMsg := simplejson.New()
	retMsg.Set("action", "del-buddy-return")
	retMsg.Set("errCode", ErrNone)
	retMsg.Set("errMsg", ErrCodeMsg[ErrNone])

	byts, _ := retMsg.MarshalJSON()
	s.SendWholeBytes(byts, false)
	return nil
}
