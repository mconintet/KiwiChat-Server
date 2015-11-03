package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mconintet/kiwi"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"math/rand"
	"time"
)

type Handler func(*simplejson.Json, kiwi.MessageSender, *kiwi.Message) error

type Router map[string]Handler

func (r Router) HandleFunc(pattern string, fn Handler) {
	r[pattern] = fn
}

func (r Router) Serve(m *kiwi.Message, s kiwi.MessageSender) error {
	j, err := simplejson.NewJson(m.Data)
	if err != nil {
		return err
	}

	action, err := j.Get("action").String()
	if err != nil {
		return errors.New("missing 'action'")
	}

	if handler, ok := r[action]; ok {
		err := handler(j, s, m)
		if err != nil {
			log.Printf("err on action: %s err: %s", action, err.Error())
		}
		return err
	}
	return errors.New("invalid action")
}

func NewRouter() Router {
	r := make(Router)
	r.HandleFunc("send-message", handleSendMessage)
	r.HandleFunc("register", handRegister)
	r.HandleFunc("login", handleLogin)
	r.HandleFunc("logout", handleLogout)
	r.HandleFunc("check-token", handleCheckToken)
	r.HandleFunc("add-buddy", handAddBuddy)
	r.HandleFunc("del-buddy", handDelBuddy)
	r.HandleFunc("get-buddies", handGetBuddies)
	r.HandleFunc("get-user-info", handleGetUserInfo)
	return r
}

const (
	MsgTypeText  = 1
	MsgTypeImage = 2
)

var (
	ErrDeformedRequest = errors.New("deformed request")
	ErrIllegalRequest  = errors.New("illegal request")
)

const (
	ErrNone            = iota
	ErrCodeUnreachable = iota + 1000
	ErrInvalidLoginInfo
	ErrEmailAlreadyExists
	ErrBuddyNotFound
	ErrBuddyAlreadyExists
	ErrInvalidToken
)

var ErrCodeMsg = map[int]string{
	ErrNone:               "",
	ErrCodeUnreachable:    "unreachable",
	ErrInvalidLoginInfo:   "invalid login info",
	ErrEmailAlreadyExists: "email already exists",
	ErrBuddyNotFound:      "buddy not found",
	ErrBuddyAlreadyExists: "buddy already exists",
	ErrInvalidToken:       "invalid token",
}

var dbConnInfo = "root:pass@unix(/tmp/mysql.sock)/kiwi"

const password_salt = "salt"

func makePassword(password string) string {
	password += password_salt
	h := md5.New()
	io.WriteString(h, password)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func makeToken() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ri := r.Int63()

	h := md5.New()
	b := make([]byte, 8)

	binary.BigEndian.PutUint64(b, uint64(ri))
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func makeDefaultAvatar() []byte {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ri := r.Uint32()

	imgColor := color.RGBA{uint8(ri >> 24), uint8(ri >> 16), uint8(ri >> 8), 255}
	img := image.NewNRGBA(image.Rect(0, 0, 300, 300))
	draw.Draw(img, img.Rect, &image.Uniform{imgColor}, image.ZP, draw.Src)

	buf := bytes.NewBuffer([]byte{})
	png.Encode(buf, img)
	return buf.Bytes()
}

func encodeAvatar(b []byte) string {
	out := bytes.NewBuffer([]byte{})
	e := base64.NewEncoder(base64.StdEncoding, out)
	e.Write(b)
	return string(out.Bytes())
}

var router = NewRouter()

func onConnOpen(r kiwi.MessageReceiver, s kiwi.MessageSender) {
	for {
		msg, err := r.ReadWhole(1 << 20)

		if err != nil {
			log.Println(err)
			s.SendClose(kiwi.CloseCodeGoingAway, "", true, false)
			break
		}

		if msg.IsText() {
			if err = router.Serve(msg, s); err != nil {
				log.Println(err)
				s.SendClose(kiwi.CloseCodeGoingAway, "", true, false)
			}
		} else if msg.IsClose() {
			s.SendClose(kiwi.CloseCodeNormalClosure, "", true, false)
			break
		}
	}
}

func onConnClose(c *kiwi.Conn) {
	uid, ok := uidConnIdMap.getUid(c.ID)
	if ok {
		resetUserNetwork(uid)
	}
	uidConnIdMap.delConnId(c.ID)
}
