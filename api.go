package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// level1 API
const (
	ApiLogin    = "https://mobile-login.xunlei.com/login"
	ApiLoginKey = "https://mobile-login.xunlei.com/loginkey"
	// ApiRenewal get, seq=uid, limitDate=exp,time_and=timestamp
	ApiRenewal  = "http://api.ext.swjsq.vip.xunlei.com/renewal?peerid=%s&sequence=%s&user_type=2&os=android-11.30RedmiK20Pro&limitdate=%s&time_and=%s&client_type=android-swjsq-2.9.3.2&sessionid=%s&client_version=androidswjsq-2.9.3.2&userid=%s&chanel=umeng-xunlei"
	ApiUserInfo = "https://mobile-login.xunlei.com/getuserinfo"
	ApiCaptcha  = "http://verify2.xunlei.com/image?t=MEA"
	ApiPortal   = "http://api.portal.swjsq.vip.xunlei.com:81/v2/queryportal"
)

// level2 API
const (
	APiCapacity = "%s/v2/bandwidth?peerid=%s"
	ApiUpgrade  = "%s/v2/upgrade?&peerid=%s&client_type=android-swjsq-2.9.3.2&client_version=androidswjsq-2.9.3.2&chanel=umeng-xunlei&os=android-11.30RedmiK20Pro&time_and=%s&userid=%s&sessionid=%s&user_type=1&needbind=1&dial_account=%s"
	// ApiKeepAlive get, portalURL,peerId,time_and=timestamp,userID,sessionID,diaACC
	ApiKeepAlive = "%s/v2/keepalive?&peerid=%s&client_type=android-swjsq-2.9.3.2&client_version=androidswjsq-2.9.3.2&chanel=umeng-xunlei&os=android-11.30RedmiK20Pro&time_and=%s&userid=%s&sessionid=%s&user_type=1&dial_account=%s"
	ApiRecover   = "%s/v2/recover?&peerid=%s&client_type=android-swjsq-2.9.3.2&client_version=androidswjsq-2.9.3.2&chanel=umeng-xunlei&os=android-11.30RedmiK20Pro&time_and=%s&userid=%s&sessionid=%s&user_type=1&dial_account=%s"
)

// API base
type API struct {
	Account *Account
	Request *Request
	Network
	Status bool
}

type Network struct {
	PortalURL   string
	DialAccount string
	//DialSequence string
}

var MacAddr string

// Session holds the api login session
type Session struct {
	ID         string
	DeviceSign string
	PeerID     string
	LoginKey   string
	UserID     string
	VipExp     string
}

func init() {
	// get mac address for device id
	var err error
	MacAddr, err = GetMacAddr()
	CheckError(err)
	MacAddr = strings.ToUpper(strings.ReplaceAll(MacAddr, ":", ""))
}

// NewAPI return new API instance
func NewAPI(account *Account) *API {
	return &API{
		Account: account,
		Request: NewRequest(),
		Status:  false,
	}
}

// GeneratePayload generate the payload from session
func (x *API) GeneratePayload(isLogin bool, isLoginKey bool) []byte {
	payload := make(map[string]string)
	// copy from common payload
	for k, v := range CommonPayload {
		payload[k] = v
	}
	// common
	payload["peerID"] = x.Account.Session.PeerID
	payload["devicesign"] = x.Account.Session.DeviceSign
	// login
	if isLogin {
		payload["passWord"] = x.Account.Password
		payload["verifyKey"] = ""
		payload["verifyCode"] = ""
		payload["isMd5Pwd"] = "0"
	}
	// login-key
	if isLoginKey {
		payload["loginKey"] = x.Account.Session.LoginKey
		payload["userName"] = x.Account.Session.UserID // this is really shitty
	} else {
		payload["sessionID"] = x.Account.Session.ID
		payload["userID"] = x.Account.Session.UserID
		payload["userName"] = strconv.FormatUint(x.Account.PhoneNumber, 10) // diff to login-key
	}
	return ForceMarshal(payload)
}

// UpdateSession key value
func (x *API) UpdateSession(id string, loginKey string) {
	x.Account.Session.LoginKey = loginKey
	x.Account.Session.ID = id
}

// FetchSessionUpdateAndSave fetch the value from api and set the session
func (x *API) FetchSessionUpdateAndSave(data map[string]interface{}) {
	// update and save session
	sID := data["sessionID"].(string)
	sLoginKey := data["loginKey"].(string)
	x.UpdateSession(sID, sLoginKey)
	x.SaveSession()
}

// SaveSession save entire Account which contains Session to the config file
func (x *API) SaveSession() {
	err := os.WriteFile(ConfigPath, ForceMarshal(x.Account), os.ModePerm)
	CheckError(err)
}

// Login generates the payload and log in
func (x *API) Login() {
	log.Println("logging you in..")
	// clean the old session
	x.Account.Session = &Session{}
	// generate fake device id and sign
	peerID := MacAddr + "004V" // first ethernet port mac address
	fakeDeviceId := GetStringMd5Hex(MacAddr)
	fakeDeviceSign := GetStringSha1Hex(fakeDeviceId + "com.xunlei.vip.swjsq68c7f21687eed3cdb400ca11fc2263c998")
	_sign := "div101." + fakeDeviceId + GetStringMd5Hex(fakeDeviceSign)
	// loads
	x.Account.Session.DeviceSign = _sign
	x.Account.Session.PeerID = peerID
	// api interaction
	payload := x.GeneratePayload(true, false)
	b, err := x.Request.Post(ApiLogin, payload)
	CheckError(err)
	// parse data
	var data map[string]interface{}
	_ = json.Unmarshal(b, &data)
	errorCode, _ := strconv.Atoi(data["errorCode"].(string))
	switch errorCode {
	case 0:
		// set userID
		x.Account.Session.UserID = data["userID"].(string)
		log.Println("login success")
	case 6:
		log.Println("MFA required!!")
		b, err := x.Request.Get(ApiCaptcha)
		CheckError(err)
		err = os.WriteFile("t.png", b, os.ModePerm)
		CheckError(err)
		log.Println("recaptcha image downloaded(./t.png)")
		Question("enter the recaptcha:")
		panic("not implemented")
		// todo: restart login with additional header
	default:
		panic("login failed")
	}
	// update and save session
	x.FetchSessionUpdateAndSave(data)
}

// LoginKey generates the payload and log in with session
func (x *API) LoginKey() {
	log.Println("logging you in with session")
	payload := x.GeneratePayload(false, true)
	b, err := x.Request.Post(ApiLoginKey, payload)
	CheckError(err)
	//parse data
	var data map[string]interface{}
	_ = json.Unmarshal(b, &data)
	errorCode, _ := strconv.Atoi(data["errorCode"].(string))
	if errorCode != 0 {
		log.Println("reload session failed, attempting to fresh login")
		x.Login()
		return
	}
	x.FetchSessionUpdateAndSave(data)
}

// GetPortal gets the network info and speedup server
func (x *API) GetPortal() {
	log.Println("getting speedup server address..")
	b, err := x.Request.Get(ApiPortal)
	CheckError(err)
	data := JsonToMap(b)
	code := data["errno"].(float64)
	if code != 0 {
		panic("get portal info failed")
	}
	province := data["province_name"].(string)
	isp := data["sp_name"].(string)
	log.Printf("ISP: %s%s\n", province, isp)
	portalIP := data["interface_ip"].(string)
	portalPort := data["interface_port"].(string)
	x.PortalURL = fmt.Sprintf("http://%s:%s", portalIP, portalPort)
}

// PromptSpeedupCapability check and prompt the speedup capability
func (x *API) PromptSpeedupCapability() {
	b, err := x.Request.Get(fmt.Sprintf(APiCapacity, x.PortalURL, MacAddr+"004V"))
	CheckError(err)
	data := JsonToMap(b)
	canUpgrade := data["can_upgrade"].(float64)
	if canUpgrade == 0 {
		panic("current network does not support the upgrade")
	}
	x.DialAccount = data["dial_account"].(string)
	currDown := int(data["bandwidth"].(map[string]interface{})["downstream"].(float64)) / 1024
	maxDown := int(data["max_bandwidth"].(map[string]interface{})["downstream"].(float64)) / 1024
	log.Println("dial account", x.DialAccount)
	log.Printf("bandwidth %dM ===> %dM\n", currDown, maxDown)
}

// CheckAccountCapability check the account vip info
func (x *API) CheckAccountCapability() {
	payload := x.GeneratePayload(true, false)
	b, err := x.Request.Post(ApiUserInfo, payload)
	CheckError(err)
	data := JsonToMap(b)
	vipList := data["vipList"]
	if vipList == nil {
		panic("none of vip found")
	}
	for _, vip := range vipList.([]interface{}) {
		t := vip.(map[string]interface{})
		vID := t["vasid"]
		vType := t["vasType"]
		if vID != nil && vType != nil && vID.(string) == "14" && vType.(string) == "2" {
			sExp := t["expireDate"].(string)
			exp, _ := time.Parse("20060102", sExp)
			if exp.Sub(time.Now()) <= 0 {
				panic("vip expired")
			}
			x.Account.Session.VipExp = sExp
			log.Println("vip expire date", sExp)
			return
		}
	}
	panic("vip not found")
}

// Renewal keep session valid
func (x *API) Renewal() bool {
	now := time.Now()
	b, err := x.Request.Get(fmt.Sprintf(ApiRenewal, x.Account.Session.PeerID, x.Account.Session.UserID,
		x.Account.Session.VipExp, strconv.FormatInt(now.UnixMilli(), 10), x.Account.Session.ID, x.Account.Session.UserID))
	CheckError(err)
	data := JsonToMap(b)
	return data["errno"].(float64) == 0
}

// AutoRenewal automatic renews the session, interval: 5min
func (x *API) AutoRenewal() {
	go func() {
		for {
			x.RetryLoginAndDo(2, x.Renewal, "upgrade")
			time.Sleep(5 * time.Minute)
		}
	}()
}

// KeepAlive the speedup
func (x *API) KeepAlive() bool {
	now := time.Now()
	b, err := x.Request.Get(fmt.Sprintf(ApiKeepAlive, x.PortalURL, x.Account.Session.PeerID, strconv.FormatInt(now.UnixMilli(), 10),
		x.Account.Session.UserID, x.Account.Session.ID, x.DialAccount))
	CheckError(err)
	data := JsonToMap(b)
	return data["errno"].(float64) == 0
}

// SpeedUp the network
func (x *API) SpeedUp() bool {
	for count := 0; count < 3; count++ {
		now := time.Now()
		b, err := x.Request.Get(fmt.Sprintf(ApiUpgrade, x.PortalURL, x.Account.Session.PeerID, strconv.FormatInt(now.UnixMilli(), 10),
			x.Account.Session.UserID, x.Account.Session.ID, x.DialAccount))
		CheckError(err)
		data := JsonToMap(b)
		errno := int(data["errno"].(float64))
		switch errno {
		case 0:
			log.Println("Upgraded???")
			return true
		case 711:
			log.Println("client request too frequent, retry after 30 min")
			time.Sleep(30 * time.Minute)
			continue
		default:
			log.Printf("unknow error code: %d => %s\n", errno, data["message"].(string))
		}
		// sleep 2 min for prevent flood detection
		log.Println("speedup failed, retry after 2min")
		time.Sleep(2 * time.Minute)
		count++
	}
	return false
}

// AutoSpeedUp automatic speedup the network, interval: 2h
func (x *API) AutoSpeedUp() {
	go func() {
		for {
			x.RetryLoginAndDo(2, x.SpeedUp, "upgrade")
			time.Sleep(2 * time.Hour)
		}
	}()
}

func (x *API) RetryLoginAndDo(max int, action func() bool, msg string) bool {
	var result bool
	// up to 2 retry
	for count := 0; count <= max; count++ {
		result = action()
		if result {
			return true
		} else {
			x.LoginKey()
		}
	}
	panic(msg + " failed")
}

// AutoKeepAlive automatic keep the speedup session, interval: 1h
func (x *API) AutoKeepAlive() {
	go func() {
		for {
			time.Sleep(3 * time.Hour)
			x.RetryLoginAndDo(2, x.KeepAlive, "keep upgraded")
		}
	}()
}

// Recover the bandwidth
func (x *API) Recover() {
	now := time.Now()
	_, err := x.Request.Get(fmt.Sprintf(ApiRecover, x.PortalURL, x.Account.Session.PeerID, strconv.FormatInt(now.UnixMilli(), 10),
		x.Account.Session.UserID, x.Account.Session.ID, x.DialAccount))
	CheckError(err)
	log.Println("bandwidth recovered")
}
