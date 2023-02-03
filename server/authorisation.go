package server

import (
	"News/domain"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/julienschmidt/httprouter"
)

type AuthCookie struct {
	ID        string `json:"id,omitempty"`
	Username  string `json:"username,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

func (s *Server) Protected(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		_, ok := s.VerifyAuthCookies(r)

		if !ok {
			fmt.Println("Check Auth - cookie error")
			http.Error(w, "cookie error -> FORBIDDEN", http.StatusForbidden)
			return
		}
		h(w, r, ps)
	}
}

func (s *Server) UserLogin(username, password, useragent string) (ac http.Cookie, err error) {
	fmt.Println("func UserLogin -> start")

	user, ex := s.ur.UserExistsInDB(username)
	if !ex {
		return ac, fmt.Errorf("login error -> user not found")
	}

	// compare passwords
	sk := os.Getenv("SECRET_KEY")
	if sk == "" {
		fmt.Println("func UserLogin -> ERROR READING SECRET_KEY -> key empty")
		log.Panic("func UserLogin -> log.PANIC -> key empty")
	}

	secretKey := []byte(sk)

	saltSum, err := base64.URLEncoding.DecodeString(user.Salt)
	if err != nil {
		return ac, fmt.Errorf("func UserLogin -> salt decode error")
	}

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(password))
	mac.Write(saltSum)
	pwSignature := mac.Sum(nil)
	passwordEnc := base64.URLEncoding.EncodeToString(pwSignature)

	fmt.Printf("User mdbs password: %v\n User form password: %v\n", user.Password, passwordEnc)
	if user.Password != passwordEnc {
		return ac, fmt.Errorf("login error => incorrect password")
	}

	ac, err = s.GenerateAuthCookie(user.Name, user.ID, useragent)

	return ac, err
}

func (s *Server) UserSave(username, password, useragent string) (ac http.Cookie, err error) {
	fmt.Println("func UserSave -> start")
	// create hmac + secret key
	sk := os.Getenv("SECRET_KEY")
	if sk == "" {
		fmt.Println("func UserSave -> ERROR READING SECRET_KEY -> key empty")
		log.Panic("func UserSave -> log.PANIC -> key empty")
	}
	secretKey := []byte(sk)
	mac := hmac.New(sha256.New, secretKey)

	// create salt
	time := time.Now().GoString()
	salt := sha256.New()
	salt.Write([]byte(time))
	saltSum := salt.Sum(nil)
	saltSumEnc := base64.URLEncoding.EncodeToString(saltSum)

	// add to hmac password and salt
	mac.Write([]byte(password))
	mac.Write(saltSum)
	pwSignature := mac.Sum(nil)
	passwordForDB := base64.URLEncoding.EncodeToString(pwSignature)

	// generate new uuid
	newUUID, err := exec.Command("uuidgen").Output()
	if err != nil {
		fmt.Println("Error create newUUID")
	}

	// user create
	user := domain.User{
		Name:     username,
		Password: passwordForDB,
		Salt:     saltSumEnc,
		ID:       string(newUUID),
	}

	fmt.Printf("User = %+v\n", user)

	// save to db
	err = s.ur.UserSave(user)
	if err != nil {
		return ac, err
	}

	ac, err = s.GenerateAuthCookie(user.Name, user.ID, useragent)

	return ac, err
}

func (s *Server) GenerateAuthCookie(username, id, useragent string) (cookie http.Cookie, err error) {
	fmt.Println("func SetAuthCookie -> start")

	// TODO save sign of user-agent as value of UserAgent
	ac := AuthCookie{
		ID:        id,
		Username:  username,
		UserAgent: useragent,
	}

	marshaledAC, err := json.Marshal(ac)
	if err != nil {
		return cookie, fmt.Errorf("func SetAuthCookie -> error Marshal cookie -> %v", err)
	}

	sk := os.Getenv("SECRET_KEY")
	if sk == "" {
		fmt.Println("func ReadAuthCookies -> ERROR READING SECRET_KEY -> key empty")
		log.Panic("func ReadAuthCookies -> log.PANIC -> key empty")
	}
	secretKey := []byte(sk)
	mac := hmac.New(sha256.New, secretKey)

	mac.Write(marshaledAC)
	expectedMAC := mac.Sum(nil)

	eVal := base64.URLEncoding.EncodeToString(append(expectedMAC, marshaledAC...))
	fmt.Println("func SetAuthCookie -> eVal -> ", eVal)

	cookie = http.Cookie{
		Name:     "auth",
		Value:    eVal,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: false,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	fmt.Println("func SetAuthCookie", cookie)
	return cookie, nil
}

func (s *Server) VerifyAuthCookies(r *http.Request) (ac AuthCookie, ok bool) {
	fmt.Println("func VerifyAuthCookies -> start")

	fmt.Printf("r.UserAgent is %+v\n", r.UserAgent())

	ac, ok = s.ReadAuthCookies(r)
	if !ok {

		return ac, false
	}

	user, ex := s.ur.UserExistsInDB(ac.Username)
	if ex && user.ID == ac.ID && ac.UserAgent == r.UserAgent() {
		return ac, true
	}
	return
}

func (s *Server) ReadAuthCookies(r *http.Request) (ac AuthCookie, ok bool) {
	fmt.Println("func ReadAuthCookies -> start")

	auth, err := r.Cookie("auth")
	if err != nil {
		fmt.Println("func ReadAuthCookies -> read auth cookie from req error: ", err)
		return ac, false
	}

	av, err := base64.URLEncoding.DecodeString(auth.Value)
	if err != nil {
		fmt.Println("func GetAuthCookies -> Decode error")
		return ac, false
	}

	err = json.Unmarshal(av[sha256.Size:], &ac)
	if err != nil {
		fmt.Println("func ReadAuthCookies -> Unmarshal error")
		return ac, false
	}

	// read curent UserAgent and add it to the cookie to compare hmac signature
	fmt.Printf("Cookie user-agent: %s\nCurent user-agent: %s\n", ac.UserAgent, r.UserAgent())
	ac.UserAgent = r.UserAgent()
	marshaledAC, err := json.Marshal(ac)
	if err != nil {
		return ac, false
	}

	// generate hmac cookie signature with curent user agent
	sk := os.Getenv("SECRET_KEY")
	if sk == "" {
		fmt.Println("func ReadAuthCookies -> ERROR READING SECRET_KEY -> key empty")
		log.Panic("func ReadAuthCookies -> log.PANIC -> key empty")
	}

	secretKey := []byte(sk)
	mac := hmac.New(sha256.New, secretKey)

	mac.Write(marshaledAC)
	expectedMAC := mac.Sum(nil)

	// compare original cookie hmac and hmac with curent user agent
	macOk := hmac.Equal(av[:sha256.Size], expectedMAC)
	fmt.Printf("macOK: %v, \n av: %v\n em: %v\n", macOk, av[:sha256.Size], expectedMAC)

	if !macOk {
		return ac, false
	}

	fmt.Println("func ReadAuthCookies -> end -> cookie ", ac.Username, ac.ID)
	return ac, true
}
